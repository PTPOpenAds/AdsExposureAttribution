/*
 * @Author: joyesjiang@tencent.com
 * @Date: 2021-01-18 00:31:43
 * @Last Modified by:   joyesjiang@tencent.com
 * @Last Modified time: 2021-01-18 00:31:43
 */

package dataprotocal

type CryptoResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    []*ResponseData `json:"data,omitempty"`
}

type ResponseData struct {
	Openid string `json:"openid"`
	//Origin string `json:"origin"`
}

func CreateErrCryptoResponse(err error) *CryptoResponse {
	return &CryptoResponse{
		Code:    -1,
		Message: err.Error(),
	}
}
