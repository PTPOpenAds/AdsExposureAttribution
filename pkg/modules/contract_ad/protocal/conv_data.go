/*
 * @Author: joyesjiang@tencent.com
 * @Date: 2021-01-15 23:03:54
 * @Last Modified by: joyesjiang@tencent.com
 * @Last Modified time: 2021-01-19 16:28:02
 */

package dataprotocal

import (
	"github.com/TencentAd/attribution/attribution/pkg/common/define"
	"github.com/golang/glog"
)

const (
	Imp               CryptoDataType = "imp"
	Conv              CryptoDataType = "conv"
	AttributionIDKey                 = "attribution_id"
	FilepathKey                      = "file_path"
	OpenidByGKey                     = "openid_g"
	ConvMatchFlag                    = "CONV_MATCHED"
	ConvMissMatchFlag                = "CONV_MISS_MATCHED"
	MessageSucc                      = "Sucess"
	FinishedDataFlag                 = "DataFinished"

	NoAttributionIDKeyError = "can't find attribution_id"
	NoFilepathKeyError      = "can't find file_path"
	NoOpenIDKeyError        = "can't find openid_g"
)

// Data : 定义输入impression
type ConvData struct {
	Openid    string `json:"openid"`
	Timestamp int    `json:"timestamp"`
}

type CryptoDataType string

type GetConvDataRequest struct {
	AttributionID string         `json:"attribution_id"`
	DataType      CryptoDataType `json:"dataType"`
}

type GetConvDataResponse struct {
	Retcode int    `json:"code"`
	Message string `json:"message"`
}

func (p *GetConvDataResponse) Init() {
	p.Retcode = 200
	p.Message = "Succ"
	glog.V(define.VLogLevel).Info("impFileProcessResponse Init: ", *p)
}

func (p *GetConvDataResponse) SetFailedResponse(errMsg string) {
	p.Retcode = 500
	p.Message = errMsg
	glog.V(define.VLogLevel).Info("impFileProcessResponse SetFailedResponse: ", *p)
}

////////////////////////////

func (p *ConvMatchResponse) Init(attributionID string) {
	p.Retcode = 200
	p.AttributionID = attributionID
	glog.V(define.VLogLevel).Info("impFileProcessResponse Init: ", *p)
}

func (p *ConvMatchResponse) SetFailedResponse(errMsg string) {
	p.Retcode = 500
	p.Message = errMsg
	glog.V(define.VLogLevel).Info("impFileProcessResponse SetFailedResponse: ", *p)
}

type ConvMatchRequest struct {
	AttributionID string      `json:"attribution_id"`
	Data          []*ConvData `json:"data"`
}

type ConvMatchResponse struct {
	Retcode       int               `json:"code"`
	Message       string            `json:"message"`
	AttributionID string            `json:"attribution_id"`
	MatchResult   map[string]string `json:"math_result"` // 拼写错误，暂时先不改
}

///////////////////////////////////////////////
/* Guest <--> Host 交互保存已匹配的Openid明文 */
///////////////////////////////////////////////

func (p *PushMatchedConvResponse) Init() {
	p.Retcode = 200
	p.Message = MessageSucc
}

func (p *PushMatchedConvResponse) SetFailedResponse(errMsg string) {
	p.Retcode = 200
	p.Message = errMsg
}

// Guest --> Host SendMatchedResult
type PushMatchedConvRequest struct {
	AttributionID string      `json:"attribution_id"`
	MatchConvInfo []*ConvData `json:"match_result"`
}

// Host --> Guest ResponseMatchedResultSaved
type PushMatchedConvResponse struct {
	Retcode int    `json:"code"`
	Message string `json:"message"`
}
