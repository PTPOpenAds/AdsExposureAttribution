package dataprotocal

import (
	"github.com/TencentAd/attribution/attribution/pkg/crypto"
)

func ProcessData(p *crypto.Parallel, groupId string, reqData *RequestData,
	cryptoFunc func(string, string) (string, error), resp *ResponseData) {

	p.AddTask(cryptoFunc, groupId, reqData.Openid, &resp.Openid)
}
