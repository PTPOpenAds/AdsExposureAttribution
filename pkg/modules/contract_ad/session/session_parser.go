/*
 * @Author: joyesjiang@tencent.com
 * @Date: 2021-01-13 11:06:42
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-01-29 15:07:16
 */

package session

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TencentAd/attribution/attribution/pkg/crypto"
	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/config"
	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/utility"
	_ "github.com/go-sql-driver/mysql"

	"github.com/TencentAd/attribution/attribution/pkg/common/define"
	"github.com/TencentAd/attribution/attribution/pkg/impression/kv"
	"github.com/golang/glog"
)

// 用于从http.request中解析对应的字段
const (
	AccountIDKey      = "account_id"
	CampaignIdsKey    = "campaign_ids"
	AppIdsKey         = "app_ids"
	TraceIDKey        = "trace_id"
	GuestConvDataPath = "guest_conv_data_path"

	noAccountIDKeyError   = "can't find account_id"
	noCampaignIdsKeyError = "can't find campaign_ids"
	noAppIdsKeyError      = "can't find app_ids"
	noTraceIDKeyError     = "can't find trace_id"
)

// CreateSessionRequest 为客户侧的请求消息结构
type CreateSessionRequest struct {
	AccountID   string `json:"account_id"`   // 广告中账户ID
	CampaignIds string `json:"campaign_ids"` // 需要归因的排期ID列表
	AppIds      string `json:"app_ids"`      // 需要归因的小程序app_id
	TraceID     string `json:"trace_id"`     // 用于上下游联调回溯的id,标记本次请求
}

// CreateSessionResponse 为客户侧的请求消息结构
type CreateSessionResponse struct {
	ReCode         int64  `json:"ret_code"`       // 当前请求处理的状态
	Message        string `json:"message"`        // 消息内容
	BigPrimeString string `json:"big_prime_hex"`  // 格式同define.Prime
	AttributionID  string `json:"attribution_id"` // 标记本次的归因的任务ID(accountid_amsid)
}

// Init 默认为处理成功
func (p *CreateSessionResponse) Init(AccountID string) {
	p.ReCode = 200
	p.Message = "Success"
	p.BigPrimeString = define.PrimeStr
	// p.BigPrimeHex = *big.NewInt(0)
	// p.AttributionID = AccountID + "_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	//p.AttributionID = strconv.FormatInt(time.Now().UnixNano(), 10)
	//p.AttributionID = strconv.FormatInt(time.Now().Unix(), 10)
	//crypto.Km.GetEncryptKey(p.AttributionID)
	glog.Infof("CreateSessionResponse Init: %v", *p)
}

// SetFailedResponse 设置请求处理失败时的信息
func (p *CreateSessionResponse) SetFailedResponse(errMsg string) {
	p.ReCode = 500
	p.Message = errMsg
	glog.Infof("CreateSessionResponse SetFailedResponse: %v ", *p)
}

func getAdgroupID(cids string) (string, string) {
	db, err := sql.Open("mysql", config.Configuration.CoreDb["connectionstr"])
	if err != nil {
		glog.Errorf("connect to db faield %v", err)
	}
	sql := config.Configuration.CoreDb["sql"] + fmt.Sprintf("('%s')", strings.Join(strings.Split(cids, ","), "','"))

	rows, err := db.Query(sql)
	if err != nil {
		glog.Errorf("Select error %+v", err)
	}
	aidSet := make(map[string]bool)
	dateSet := make(map[string]bool)

	for rows.Next() {
		var aid, date string
		rows.Scan(&aid, &date)
		aidSet[aid] = true
		// 2021-01-14 00:00:00 ->  20210114
		dateSet[strings.Replace(strings.Split(date, " ")[0], "-", "", -1)] = true
	}
	var aidList []string
	var dateList []string
	for k := range aidSet {
		aidList = append(aidList, k)
	}
	for k := range dateSet {
		dateList = append(dateList, k)
	}

	return strings.Join(aidList, ","), strings.Join(dateList, ",")
}

// CallCollectImpData 执行脚本收集CampaignId对应的曝光数据并转化为OpenId
func CallCollectImpData(cids string, appids string, attrID string) (b bool, err error) {
	adgroupIDs, dates := getAdgroupID(cids)
	if err := utility.RunCommandWithRetry(utility.LoadExpData(dates, adgroupIDs, attrID)); err != nil {
		return false, err
	}
	if err := utility.RunCommandWithRetry(utility.DumpData(attrID)); err != nil {
		return false, err
	}
	return true, nil
}

func calcAttributionID(cids string, appids string) string {
	idList := append(strings.Split(cids, ","), strings.Split(appids, ",")...)
	sort.Strings(idList)
	md5Hex := utility.MD5(strings.Join(idList, ","))[:15]
	intV, _ := strconv.ParseInt(md5Hex, 16, 64)
	return strconv.FormatInt(intV, 10)
}

type sessionHandler struct {
	kv     kv.KV
	jobMap map[string]bool
}

// NewSessionHandler 处理外部Createsession请求
func NewSessionHandler(kv kv.KV) *sessionHandler {
	return &sessionHandler{
		kv:     kv,
		jobMap: make(map[string]bool),
	}
}

type guestSessionHandler struct {
	kv              kv.KV
	hostSessionPath string
}

func NewGuestSessionHandler(kv kv.KV, hostSessionPath string) *guestSessionHandler {
	return &guestSessionHandler{
		kv:              kv,
		hostSessionPath: hostSessionPath,
	}
}

func GuestDataPathKey(attributionId string) string {
	return attributionId + "_GuestSession"
}

func (s *guestSessionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer glog.Flush()
	defer r.Body.Close()

	values := r.URL.Query()
	glog.Info("Display_req: ", values)
	// sessionRequest 其实用不到，因为是Get请求，可以直接拿参数

	if values.Get(AccountIDKey) == "" || values.Get(CampaignIdsKey) == "" ||
		values.Get(AppIdsKey) == "" || values.Get(GuestConvDataPath) == "" {
		_, _ = w.Write([]byte("some params missed"))
		return
	}

	paramsMap := make(map[string]string)
	paramsMap["account_id"] = values.Get(AccountIDKey)
	paramsMap["campaign_ids"] = values.Get(CampaignIdsKey)
	paramsMap["app_ids"] = values.Get(AppIdsKey)
	paramsMap["trace_id"] = strconv.FormatInt(time.Now().UnixNano(), 10)
	var sessResp CreateSessionResponse

	err := utility.DoHTTPGet(s.hostSessionPath, utility.ContentType, &paramsMap, &sessResp)
	if err != nil {
		glog.Errorf("Do http get failed! : %s \n", err)
	}

	crypto.Km.GetEncryptKey(sessResp.AttributionID)
	glog.Infof("Create Session Response: %+v\n", sessResp)

	if err := s.kv.Set(GuestDataPathKey(sessResp.AttributionID), values.Get(GuestConvDataPath)); err != nil {
		glog.Errorf("Init Guest Session failed : %s \n", err)
	}

}

func (s *sessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	glog.Info("Display_req: ", req)
	errHandle := func(err error) {
		log.Print(err)
		_, _ = w.Write([]byte(err.Error()))
	}

	values := req.URL.Query()
	glog.Info("Display_req: ", values)

	var sessReq CreateSessionRequest
	var sessResp CreateSessionResponse

	sessReq.AccountID = string(values.Get(AccountIDKey))
	if sessReq.AccountID == "" {
		errHandle(fmt.Errorf(noAccountIDKeyError))
		sessResp.Message = noAccountIDKeyError
		return
	}

	sessReq.CampaignIds = string(values.Get(CampaignIdsKey))
	if sessReq.CampaignIds == "" {
		errHandle(fmt.Errorf(noCampaignIdsKeyError))
		sessResp.Message = noCampaignIdsKeyError
		return
	}

	sessReq.AppIds = string(values.Get(AppIdsKey))
	if sessReq.AppIds == "" {
		errHandle(fmt.Errorf(noAppIdsKeyError))
		sessResp.Message = noAppIdsKeyError
		return
	}

	sessReq.TraceID = values.Get(TraceIDKey)
	if sessReq.TraceID == "" {
		errHandle(fmt.Errorf(noTraceIDKeyError))
		sessResp.Message = noTraceIDKeyError
	}

	// 防止重发
	sessResp.AttributionID = calcAttributionID(sessReq.CampaignIds, sessReq.AppIds)
	if _, ok := s.jobMap[sessResp.AttributionID]; ok {
		errHandle(fmt.Errorf("duplicated attributionid %s", values))
		return
	}
	s.jobMap[sessResp.AttributionID] = true

	crypto.Km.GetEncryptKey(sessResp.AttributionID)
	glog.Infof("Calculate attribution id, appids:%s, campaign_ids:%s, attributionid:%s",
		sessReq.AppIds, sessReq.CampaignIds, sessResp.AttributionID)

	var key = sessReq.AccountID + "_CreateSessionRequest"
	data, _ := json.Marshal(sessReq)
	if err := s.kv.Set(key, string(data)); err != nil {
		errHandle(err)
	}

	sessResp.Init(sessReq.AccountID)
	key = sessReq.AccountID + "_CreateSessionResponse"
	data, _ = json.Marshal(sessResp)
	if err := s.kv.Set(key, string(data)); err != nil {
		errHandle(err)
	}
	_, _ = w.Write(data)
	glog.Info("Display_response: ", sessResp)

	glog.Flush()
	go func() {
		if ok, err := CallCollectImpData(sessReq.CampaignIds, sessReq.AppIds, sessResp.AttributionID); !ok {
			sessResp.SetFailedResponse("Failed Do CallCollectImpData")
			glog.Info("Failed Do CallCollectImpData, req: ", req, "err: ", err)
		} else {
			sendRequest := fmt.Sprintf("curl -XGET 'http://127.0.0.1:9081/host/imp_process?attribution_id=%s&app_ids=%s'", sessResp.AttributionID, sessReq.AppIds)
			glog.Infof("[INFO] SendRequest: %s", sendRequest)
			utility.RunCommandRealTime(sendRequest)
		}
	}()
}
