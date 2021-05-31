/*
 * @Author: joyesjiang@tencent.com
 * @Date: 2021-01-15 23:03:54
 * @Last Modified by: joyesjiang@tencent.com
 * @Last Modified time: 2021-01-20 19:29:05
 */

package dataprotocal

type CryptoRequest struct {
	AttributionID string         `json:"attribution_id"`
	Data          []*RequestData `json:"data"`
	DataType      CryptoDataType `json:"dataType"`
}

type RequestData struct {
	Openid string `json:"openid"`
}
