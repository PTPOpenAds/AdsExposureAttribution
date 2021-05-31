/*
 * @Author: joyesjiang@tencent.com
 * @Date: 2021-01-15 23:01:44
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-01-28 20:37:34
 */

package dataprocess

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/TencentAd/attribution/attribution/pkg/common/define"
	"github.com/TencentAd/attribution/attribution/pkg/common/metricutil"
	"github.com/TencentAd/attribution/attribution/pkg/crypto"
	"github.com/TencentAd/attribution/attribution/pkg/handler/http/encrypt/metrics"
	"github.com/TencentAd/attribution/attribution/pkg/handler/http/encrypt/safeguard"
	dataprotocal "github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/protocal"
	"github.com/golang/glog"
)

type HttpHandle struct {
	convEncryptSafeguard *safeguard.ConvEncryptSafeguard
}

func NewCryptoHandle() *HttpHandle {
	return &HttpHandle{}
}

func (handle *HttpHandle) WithSafeguard(guard *safeguard.ConvEncryptSafeguard) *HttpHandle {
	handle.convEncryptSafeguard = guard
	return handle
}

func (handle *HttpHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	var err error
	defer func() {
		metricutil.CollectMetrics(metrics.ConvEncryptErrCount, metrics.ConvEncryptHandleCost, startTime, err)
	}()

	var resp *dataprotocal.CryptoResponse
	if resp, err = handle.doServeHttp(r); err != nil {
		resp = dataprotocal.CreateErrCryptoResponse(err)
	}

	handle.writeResponse(w, resp)
}

func (handle *HttpHandle) doServeHttp(r *http.Request) (*dataprotocal.CryptoResponse, error) {
	var err error
	var body []byte
	body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	glog.V(define.VLogLevel).Infof("ReceiveCryptoRequest: %s", string(body))
	if glog.V(define.VLogLevel) {
		glog.V(define.VLogLevel).Infof("ReceiveCryptoRequest: %s", string(body))
	}

	var req dataprotocal.CryptoRequest
	if err = json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	intGroupID, _ := strconv.ParseInt(req.AttributionID, 10, 64)
	err = handle.convEncryptSafeguard.Against(intGroupID)
	if err != nil {
		return nil, err
	}

	glog.V(define.VLogLevel).Info("Crypto_doServeHttp: ", intGroupID, req)
	groupID := req.AttributionID
	resp := &dataprotocal.CryptoResponse{
		Message: "success",
	}

	for _, reqData := range req.Data {
		var respData dataprotocal.ResponseData
		respData.Openid, err = crypto.Encrypt(groupID, reqData.Openid)
		if err != nil {
			glog.Infof("Guest encrypted failed, reqData:%v\n", reqData)
		}
		resp.Data = append(resp.Data, &respData)
	}

	rand.Shuffle(len(resp.Data), func(i, j int) {
		resp.Data[i], resp.Data[j] = resp.Data[j], resp.Data[i]
	})

	glog.V(define.VLogLevel).Info("Crypto_doServeHttp: ", resp, len(resp.Data))

	return resp, nil
}

func (handle *HttpHandle) writeResponse(w http.ResponseWriter, resp *dataprotocal.CryptoResponse) {
	data, err := json.Marshal(resp)
	if glog.V(define.VLogLevel) {
		glog.V(define.VLogLevel).Infof("encrypt response: %s", string(data))
	}
	if err != nil {
		w.WriteHeader(500)
	} else {
		_, _ = w.Write(data)
	}
}
