/*
 * @Author: joyesjiang@tencent.com
 * @Date: 2021-01-17 22:58:03
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-01-28 19:05:06
 */

package utility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/TencentAd/attribution/attribution/pkg/common/define"
	"github.com/golang/glog"
)

var (
	// ContentType  -- Http请求格式描述
	ContentType = "application/json;charset=utf-8"
)

const retryTimes = 5

// DoHTTPPost 发送http Post请求
func DoHTTPPost(postURL string, contType string, reqData interface{}, respData interface{}) error {
	reqJSON, err := json.Marshal(reqData)
	if err != nil {
		return err
	}
	if postURL == "" {
		return fmt.Errorf("DoHTTPPost failed: %s", postURL)
	}

	glog.V(define.VLogLevel).Infof("DoHTTPPost reqJSON: %v", reqJSON)
	// 外发Post请求
	var result *http.Response
	// 重试retryTimes
	for i := 0; i < retryTimes; i++ {
		result, err = http.Post(postURL, contType, bytes.NewBuffer([]byte(reqJSON)))
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return err
	}
	json.Unmarshal([]byte(body), respData)
	glog.V(define.VLogLevel).Infof("DoHTTPPost respJSON: %v", respData)

	return nil
}

// DoHTTPGet 发送http Get请求
func DoHTTPGet(getURL string, contType string, paramsMap *map[string]string, respData interface{}) error {
	params := url.Values{}
	url, err := url.Parse(getURL)

	for key, value := range *paramsMap {
		params.Set(key, value)
	}

	url.RawQuery = params.Encode()
	urlPath := url.String()
	glog.V(define.VLogLevel).Infof("Send Get Request: %s", urlPath)
	var resp *http.Response
	for i := 0; i < retryTimes; i++ {
		resp, err = http.Get(urlPath)
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(body), &respData)
	if err != nil {
		return err
	}
	return nil
}
