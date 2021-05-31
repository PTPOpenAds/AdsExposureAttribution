/*
 * @Author: joyesjiang@tencent.com
 * @Date: 2021-01-17 21:20:47
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-01-24 13:34:25
 */

package dataprocess

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/utility"

	"github.com/TencentAd/attribution/attribution/pkg/common/define"
	dataprotocal "github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/protocal"

	"github.com/golang/glog"
)

type saveInfo struct {
	ch  chan string
	cnt uint64
}

// SaveMatchResultHandler 用于处理曝光文件数据的Handler
type SaveMatchResultHandler struct {
	saveInfoMap             map[string]*saveInfo
	hostMatchResultSavePath string
	sync.Mutex
}

// NewHostSaveMatchResultHandler : 构造Handler
func NewHostSaveMatchResultHandler(hostMatchResultSavePath string) *SaveMatchResultHandler {
	glog.Info("NewImpressionFileHandle.")
	newHandler := SaveMatchResultHandler{
		// Host将所有的匹配记录传入Channel中输出到文件
		saveInfoMap:             make(map[string]*saveInfo),
		hostMatchResultSavePath: hostMatchResultSavePath,
	}
	//go utility.WriteToFile(newHandler.ch, hostMatchResultSavePath+"."+time.Now().Format("20060102150405"))
	return &newHandler
}

func (handle *SaveMatchResultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	errHandle := func(err error) {
		log.Print(err)
		_, _ = w.Write([]byte(err.Error()))
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errHandle(err)
		return
	}
	if glog.V(define.VLogLevel) {
		glog.V(define.VLogLevel).Infof("SaveMatchResultHandler body: %s", string(body))
	}

	var saveMatchReq dataprotocal.PushMatchedConvRequest
	var saveMatchResp dataprotocal.PushMatchedConvResponse
	saveMatchResp.Init()
	if err = json.Unmarshal(body, &saveMatchReq); err != nil {
		return
	}

	// 对于map的读写要加锁
	handle.Lock()
	if _, ok := handle.saveInfoMap[saveMatchReq.AttributionID]; !ok {
		ch := make(chan string, 100)
		handle.saveInfoMap[saveMatchReq.AttributionID] = &saveInfo{
			ch:  ch,
			cnt: 0,
		}
		go utility.WriteToFile(ch, handle.hostMatchResultSavePath+"/"+saveMatchReq.AttributionID)
	}
	curSaveInfo := handle.saveInfoMap[saveMatchReq.AttributionID]
	handle.Unlock()

	// Host将所有的匹配记录传入Channel中,输出到文件
	// Host这里来控制当收到的CovData中OpenId=FinishedDataFlags时,关闭Channel
	for _, convData := range saveMatchReq.MatchConvInfo {
		if convData.Openid != dataprotocal.FinishedDataFlag {
			atomic.AddUint64(&curSaveInfo.cnt, 1)
			stringData, _ := json.Marshal(*convData)
			select {
			case curSaveInfo.ch <- string(stringData):
			case <-time.After(1 * time.Second):
				glog.Error("SaveMatchResultHandler Failed: ", convData.Openid)
			}
		} else {
			glog.V(define.VLogLevel).Info("SaveMatchResultHandler Finished, totalNum: ", curSaveInfo.cnt, ", req: ", saveMatchReq)
			close(curSaveInfo.ch)
		}
	}

	data, err := json.Marshal(saveMatchResp)
	if glog.V(define.VLogLevel) {
		glog.V(define.VLogLevel).Infof("SaveMatchResultHandler MatchedOpenid: %s", string(data))
	}
	if err != nil {
		w.WriteHeader(500)
	} else {
		_, _ = w.Write(data)
	}
}
