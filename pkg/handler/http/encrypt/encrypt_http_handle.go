/*
 * copyright (c) 2020, Tencent Inc.
 * All rights reserved.
 *
 * Author:  linceyou@tencent.com
 * Last Modify: 11/3/20, 2:26 PM
 */

package encrypt

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
	"github.com/TencentAd/attribution/attribution/pkg/crypto/protocal"
	"github.com/TencentAd/attribution/attribution/pkg/handler/http/encrypt/metrics"
	"github.com/TencentAd/attribution/attribution/pkg/handler/http/encrypt/safeguard"
	"github.com/golang/glog"
)

type HttpHandle struct {
	convEncryptSafeguard *safeguard.ConvEncryptSafeguard
}

func NewHttpHandle() *HttpHandle {
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

	var resp *protocal.CryptoResponse
	if resp, err = handle.doServeHttp(r); err != nil {
		resp = protocal.CreateErrCryptoResponse(err)
	}

	handle.writeResponse(w, resp)
}

func (handle *HttpHandle) doServeHttp(r *http.Request) (*protocal.CryptoResponse, error) {
	var err error
	var body []byte
	body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if glog.V(define.VLogLevel) {
		glog.V(define.VLogLevel).Infof("encrypt body: %s", string(body))
	}

	var req protocal.CryptoRequest
	if err = json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	err = handle.convEncryptSafeguard.Against(req.CampaignId)
	if err != nil {
		return nil, err
	}

	groupId := strconv.FormatInt(req.CampaignId, 10)
	resp := &protocal.CryptoResponse{
		Message:    "success",
	}
	var p crypto.Parallel
	for _, reqData := range req.Data {
		var respData protocal.ResponseData
		protocal.ProcessData(&p, groupId, reqData, crypto.Encrypt, &respData)
		resp.Data = append(resp.Data, &respData)
	}
	if err = p.WaitAndCheck(); err != nil {
		return nil, err
	}

	rand.Shuffle(len(resp.Data), func(i, j int) {
		resp.Data[i], resp.Data[j] = resp.Data[j], resp.Data[i]
	})

	return resp, nil
}

func (handle *HttpHandle) writeResponse(w http.ResponseWriter, resp *protocal.CryptoResponse) {
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
