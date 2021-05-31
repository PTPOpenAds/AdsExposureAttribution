/*
 * @Author: joyesjiang@tencent.com
 * @Date: 2021-01-17 21:20:47
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-01-28 19:03:33
 */

package dataprocess

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/TencentAd/attribution/attribution/pkg/impression/kv"
	"github.com/go-redis/redis/v8"

	"github.com/TencentAd/attribution/attribution/pkg/common/define"
	dataProtocal "github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/protocal"

	"github.com/golang/glog"
)

// ConvMatchHandler : 定义处理曝光handler
type ConvMatchHandler struct {
	attributionID string
	kv            kv.KV
}

// MatchProcessHandler 用于处理曝光文件数据的Handler
type MatchProcessHandler struct {
	matchHandler *ConvMatchHandler
}

// NewHostConvMatchHandler : 构造 ConvMatchHandler
func NewHostConvMatchHandler(kv kv.KV) *MatchProcessHandler {
	glog.Info("NewHostConvMatchHandler.")
	return &MatchProcessHandler{
		matchHandler: &ConvMatchHandler{
			kv: kv,
		},
	}
}

func (handle *MatchProcessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	errHandle := func(err error) {
		log.Print(err)
		_, _ = w.Write([]byte(err.Error()))
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	if glog.V(define.VLogLevel) {
		glog.V(define.VLogLevel).Infof("SaveMatchResultHandler body: %s", string(body))
	}

	var convMatchReq dataProtocal.ConvMatchRequest
	if err = json.Unmarshal(body, &convMatchReq); err != nil {
		return
	}
	var convMatchResp dataProtocal.ConvMatchResponse
	handle.matchHandler.attributionID = convMatchReq.AttributionID

	if convMatchReq.AttributionID == "" {
		errHandle(fmt.Errorf(dataProtocal.NoAttributionIDKeyError))
		convMatchResp.SetFailedResponse(dataProtocal.NoAttributionIDKeyError)
		return
	}
	convMatchResp.Init(convMatchReq.AttributionID)

	if len(convMatchReq.Data) == 0 {
		errHandle(fmt.Errorf(dataProtocal.NoOpenIDKeyError))
		convMatchResp.SetFailedResponse(dataProtocal.NoOpenIDKeyError)
		return
	}

	// 将是否matched构建成map返回
	convMatchResp.MatchResult = make(map[string]string)
	for _, convData := range convMatchReq.Data {
		if value, err := handle.matchHandler.kv.Get(ConstructKey(handle.matchHandler.attributionID, convData.Openid)); err == redis.Nil {
			convMatchResp.MatchResult[convData.Openid] = dataProtocal.ConvMissMatchFlag
			glog.Infof("MatchProcessHandler MissMatched, openid: %s, attributionid:%s, result:%s\n", convData.Openid, handle.matchHandler.attributionID, value)
		} else {
			convMatchResp.MatchResult[convData.Openid] = dataProtocal.ConvMatchFlag
			glog.Infof("MatchProcessHandler MatchSucc, openid: %s, attributionid:%s, result:%s\n", convData.Openid, handle.matchHandler.attributionID, value)
		}
	}

	data, err := json.Marshal(convMatchResp)
	if glog.V(define.VLogLevel) {
		glog.V(define.VLogLevel).Infof("MatchProcessHandler response: %s", string(data))
	}
	if err != nil {
		w.WriteHeader(500)
	} else {
		_, _ = w.Write(data)
	}
}
