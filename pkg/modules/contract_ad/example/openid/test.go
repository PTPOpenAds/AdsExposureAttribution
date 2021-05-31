/*
 * copyright (c) 2020, Tencent Inc.
 * All rights reserved.
 *
 * Author:  linceyou@tencent.com
 * Last Modify: 11/13/20, 10:16 AM
 */

package main

import (
	"flag"
	"fmt"
	"strings"
	"sync"

	"git.code.oa.com/components/l5"
	"github.com/TencentAd/attribution/attribution/pkg/common/flagx"
	dataprocess "github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/data_process"
	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/utility"
)

var (
	dataPath = flag.String("data_path", "data", "data_path")
)

func main() {
	flagx.Parse()

	l5Api, _ := l5.NewDefaultApi()
	var wuidWG sync.WaitGroup

	fileList := utility.GetFileList(*dataPath)

	// 1. 121244488 wx25f982a55e60a540
	// 2. 50267459 wx4ed9e1f4e0f3eeb0
	// 3. 11469945 wx5b7e52da9b9c8ade

	for _, filePath := range fileList {
		if !utility.Exists(filePath + "_openid" + utility.DefaultSuffix) {

			filePathList := strings.Split(filePath, "/")
			if strings.Contains(filePathList[len(filePathList)-1], "_") {
				continue
			}

			fmt.Println(filePath)
			wuidHandle := dataprocess.NewWuidHandler(filePath, "12345", []string{"wx25f982a55e60a540", "wx4ed9e1f4e0f3eeb0", "wx5b7e52da9b9c8ade"},
				l5Api, &wuidWG)
			wuidHandle.Run()
			wuidWG.Wait()
		}
	}
}
