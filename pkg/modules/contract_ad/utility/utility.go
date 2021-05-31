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
)

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
