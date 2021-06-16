/*
 * @Author: joyesjiang@tencent.com
 * @Date: 2021-01-17 21:20:47
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-01-29 17:01:00
 */

package dataprocess

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/TencentAd/attribution/attribution/pkg/common/define"
	"github.com/TencentAd/attribution/attribution/pkg/crypto"
	"github.com/TencentAd/attribution/attribution/pkg/crypto/util"
	"github.com/TencentAd/attribution/attribution/pkg/impression/kv"
	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/config"
	dataprotocal "github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/protocal"
	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/session"
	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/utility"
	"github.com/golang/glog"
)

var (
	isSendConvDetail = flag.Bool("is_send_conv_detail", true, "whether send conv detail info")

	// 统计总的请求行数
	lineNums uint64 = 0
)

// fileHandle : 定义处理曝光handler
type fileHandle struct {
	filepath    string
	convHandler ConvProcessHandler
	// Guest 侧存储所有匹配成功的OpenID信息
	ch chan string
	kv kv.KV
}

// ConvProcessHandler 用于处理曝光文件数据的Handler
type ConvProcessHandler struct {
	// Guest侧保存Conv数据的文件路径
	attributionID string

	// Guest --> Host 发送g(openid_conv)到Host进行数据匹配
	hostConvMatchPushPath string
	// Guest --> Host 推送最终匹配的conv对应的openid明文
	hostSaveMatchedPushPath string
	// Guest 侧存储所有匹配成功的OpenID信息
	guestMatchResultSavePath string

	kv kv.KV
}

// NewConvProcessFileHandle : Guest Process ConvData Handler
func NewConvProcessFileHandle(guestMatchResultSavePath string,
	hostConvMatchPushSvr string, hostSaveMatchedPushSvr string, kv kv.KV) *ConvProcessHandler {
	glog.Info("NewImpressionFileHandle guestMatchResultSavePath: ", guestMatchResultSavePath)
	glog.Info("NewImpressionFileHandle host_conv_match push_path: ", hostConvMatchPushSvr)
	glog.Info("NewImpressionFileHandle host_save_match push_path: ", hostSaveMatchedPushSvr)
	return &ConvProcessHandler{
		hostConvMatchPushPath:    hostConvMatchPushSvr,
		hostSaveMatchedPushPath:  hostSaveMatchedPushSvr,
		guestMatchResultSavePath: guestMatchResultSavePath,
		kv:                       kv,
	}
}

func (handle *ConvProcessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	errHandle := func(err error) {
		log.Print(err)
		_, _ = w.Write([]byte(err.Error()))
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("parse body failed! err:%s", err)
		return
	}

	glog.Infof("ConvProcessHandler body: %s", string(body))

	var convReq dataprotocal.GetConvDataRequest
	var convResp dataprotocal.GetConvDataResponse
	convResp.Init()
	if err = json.Unmarshal(body, &convReq); err != nil {
		glog.Errorf("unmarshal body failed! err:%s, body:%s", err, body)
		return
	}
	glog.Info("Display_ConvProcessHandler req: ", convReq)

	if convReq.AttributionID == "" {
		errHandle(fmt.Errorf(dataprotocal.NoAttributionIDKeyError))
		convResp.SetFailedResponse(dataprotocal.NoAttributionIDKeyError)
		return
	}

	handle.attributionID = convReq.AttributionID
	fp, err := handle.kv.Get(session.GuestDataPathKey(convReq.AttributionID))
	if err != nil {
		glog.Errorf("get file path error")
	}
	newHandler := fileHandle{
		convHandler: *handle,
		ch:          make(chan string),
		filepath:    fp,
	}
	glog.Info("Ready_to_run, conv_file_path: ", newHandler.filepath)
	defer func() {
		glog.Infof("StartRunFileHandle: %s", newHandler.filepath)
		go utility.WriteToFile(newHandler.ch, newHandler.convHandler.guestMatchResultSavePath+"/"+
			handle.attributionID+"_"+time.Now().Format(define.YYYYMMDD))
		newHandler.Run()
	}()

	data, err := json.Marshal(convResp)
	glog.Infof("encrypt response: %s", string(data))

	if err != nil {
		w.WriteHeader(500)
	} else {
		_, _ = w.Write(data)
	}
}

// Run : Guest加载Conv文件数据
func (p *fileHandle) Run() error {
	glog.Info("fileHandle.Run", p.filepath)
	lp := utility.NewLineProcess(p.filepath, p.processLine, func(line string, err error) {
		glog.Errorf("failed to handle line[%s], err[%v]", line, err)
	})

	if err := lp.LoadFile(*utility.DefaultNlines, utility.DefaultSeparator, utility.DefaultSuffix); err != nil {
		return err
	}

	lp.WaitDone()
	saveMatchReq := dataprotocal.PushMatchedConvRequest{
		AttributionID: p.convHandler.attributionID,
	}
	saveMatchReq.MatchConvInfo = append(saveMatchReq.MatchConvInfo, &dataprotocal.ConvData{
		Openid: dataprotocal.FinishedDataFlag,
	})
	glog.Info("Finished fileHandle Run: ", p.filepath, ", total_lineNums:", lineNums, ", req: ", saveMatchReq)

	close(p.ch)
	var saveMatchResp dataprotocal.PushMatchedConvResponse
	utility.DoHTTPPost(p.convHandler.hostSaveMatchedPushPath, utility.ContentType, saveMatchReq, &saveMatchResp)
	glog.Info("All process has been Done!")
	glog.Flush()
	//os.Exit(0)
	return nil
}

func constructConvInfo(line string, attributionID string) *dataprotocal.ConvData {
	accountConf, ok := config.Configuration.AccountConfMap[session.AttributionAccountMap[attributionID]]
	if ok && accountConf.DataType == "txt" {
		openid := strings.Split(line, accountConf.Seperator)[accountConf.OpenidIndex]
		return &dataprotocal.ConvData{
			Openid: openid,
			Detail: line,
		}
	} else {
		var convInfo dataprotocal.ConvData
		json.Unmarshal([]byte(line), &convInfo)
		return &convInfo
	}
}

func (p *fileHandle) processLine(lines string) error {

	lineList := strings.Split(lines, utility.DefaultSeparator)
	convMatchReq := &dataprotocal.ConvMatchRequest{
		AttributionID: p.convHandler.attributionID,
	}
	var convInfoMap = make(map[string]*dataprotocal.ConvData)

	// 加密发送请求数据
	for _, line := range lineList {
		convInfo := constructConvInfo(line, p.convHandler.attributionID)
		atomic.AddUint64(&lineNums, 1)
		if lineNums%10000 == 0 {
			p.convHandler.kv.Set(p.convHandler.attributionID+"::GuestConvProcessNum", strconv.Itoa(int(lineNums)))
		}

		glog.V(define.VLogLevel).Info("processLine: ", convInfo, line)

		ImpFOpenid, err := crypto.Encrypt(p.convHandler.attributionID, util.StringToHex(convInfo.Openid))
		convInfoMap[ImpFOpenid] = convInfo
		if err != nil {
			return err
		}
		glog.V(define.VLogLevel).Info("processLine ImpFOpenid: ", ImpFOpenid)

		convMatchReq.Data = append(convMatchReq.Data, &dataprotocal.ConvData{Openid: ImpFOpenid})
	}

	// 发送请求
	var convMatchResp dataprotocal.ConvMatchResponse
	if err := utility.DoHTTPPost(p.convHandler.hostConvMatchPushPath, utility.ContentType, convMatchReq, &convMatchResp); err != nil {
		glog.Infof("Guest Send Conv Data Failed %+v", convMatchReq)
	}

	// 解析返回值确定是否要发送明文
	saveMatchReq := dataprotocal.PushMatchedConvRequest{
		AttributionID: p.convHandler.attributionID,
	}

	var saveMatchResp dataprotocal.PushMatchedConvResponse
	for fopenid, isMatched := range convMatchResp.MatchResult {
		if isMatched == dataprotocal.ConvMatchFlag {
			stingData, _ := json.Marshal(*convInfoMap[fopenid])
			p.ch <- string(stingData)
			saveMatchReq.MatchConvInfo = append(saveMatchReq.MatchConvInfo, convInfoMap[fopenid])
		}
	}

	// 增加是否发送明文转化的开关
	if *isSendConvDetail && len(saveMatchReq.MatchConvInfo) > 0 {
		utility.DoHTTPPost(p.convHandler.hostSaveMatchedPushPath, utility.ContentType, saveMatchReq, &saveMatchResp)
		return nil
	}

	//glog.V(define.VLogLevel).Info("processLine err: ", convMatchResp.MatchResult, ", ", dataprotocal.ConvMatchFlag)
	return nil
}
