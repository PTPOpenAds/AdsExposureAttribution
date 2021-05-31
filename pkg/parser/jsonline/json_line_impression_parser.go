package jsonline

import (
	"encoding/json"

	"github.com/TencentAd/attribution/attribution/proto/impression"
)

// 按照json格式解析http的body
type JsonImpressionParser struct{}

func NewJsonImpressionParser() *JsonImpressionParser {
	return &JsonImpressionParser{}
}

func (p *JsonImpressionParser) Parse(input interface{}) (*impression.Request, error) {
	line := input.(string)
	request := new(impression.Request)
	err := json.Unmarshal([]byte(line), request)
	if err != nil {
		return nil, err
	}
	return request, nil
}
