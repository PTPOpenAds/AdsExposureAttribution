package parser

import (
	"flag"
	"fmt"
	"strings"

	"github.com/TencentAd/attribution/attribution/pkg/parser/jsonline"
	"github.com/TencentAd/attribution/attribution/proto/impression"
)

var (
	impressionParserName = flag.String("impression_parser_name", "ams", "")
)

type ImpressionParserInterface interface {
	Parse(input interface{}) (*impression.Request, error)
}

func CreateImpressionParser() (ImpressionParserInterface, error) {
	switch strings.ToLower(*impressionParserName) {
	// case "ams":
	// 	return ams.NewAMSClickParser(), nil
	case "jsonline":
		return jsonline.NewJsonImpressionParser(), nil

	default:
		return nil, fmt.Errorf("impression parser [%s] not support", *impressionParserName)
	}
}
