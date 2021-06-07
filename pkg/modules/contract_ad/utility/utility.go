/*
 * @Author: xiaoyangma@tencent.com
 * @Date: 2021-01-26 21:34:31
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-01-26 21:35:15
 */

package utility

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func CalcAttributionID(cids string, appids string) string {
	idList := append(strings.Split(cids, ","), strings.Split(appids, ",")...)
	sort.Strings(idList)
	md5Hex := MD5(strings.Join(idList, ","))[:15]
	intV, _ := strconv.ParseInt(md5Hex, 16, 64)
	return strconv.FormatInt(intV, 10)
}
