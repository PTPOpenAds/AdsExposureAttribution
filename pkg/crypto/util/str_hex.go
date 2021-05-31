package util

import (
	"encoding/hex"
)

func StringToHex(str string) string {
	return hex.EncodeToString([]byte(str))
}

func HexToString(str string) string {
	output, _ := hex.DecodeString(str)
	return string(output)
}
