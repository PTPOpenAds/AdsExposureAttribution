/*
 * copyright (c) 2020, Tencent Inc.
 * All rights reserved.
 *
 * Author:  linceyou@tencent.com
 * Last Modify: 11/13/20, 10:16 AM
 */

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/TencentAd/attribution/attribution/pkg/common/flagx"
	"github.com/TencentAd/attribution/attribution/pkg/impression/kv"
	"github.com/TencentAd/attribution/attribution/pkg/impression/kv/opt"
)

var (
	inputFile        = flag.String("inputFile", "input.dat", "")
	kvType           = flag.String("kv_type", "HASH_LEVELDB", "")
	kvAddress        = flag.String("kv_address", "./db", "")
	kvPassword       = flag.String("kv_password", "", "")
	gStorage   kv.KV = nil

// outputFile    = flag.String("outputFile", "output.dat", "")
// isEncrypt     = flag.Bool("isEncrypt", true, "isEncrypt")
// groupID       = flag.String("groupID", "123456", "groupID")
// isFirstInput  = flag.Bool("isFirstInput", false, "isFirstInput")
// isFinalOutput = flag.Bool("isFinalOutput", false, "isFinalOutput")
)

/*
go run crypto_example.go --default_redis_config="{\"address\":[\"127.0.0.1:6379\"]}" --key_manager_name=redis --inputFile=input.dat --outputFile=fx.dat --isEncrypt=true --isFirstInput=true --groupID=123456
go run crypto_example.go --default_redis_config="{\"address\":[\"127.0.0.1:6379\"]}" --key_manager_name=redis --inputFile=fx.dat --outputFile=gfx.dat --isEncrypt=true --groupID=1234567
go run crypto_example.go --default_redis_config="{\"address\":[\"127.0.0.1:6379\"]}" --key_manager_name=redis --inputFile=gfx.dat --outputFile=gx.dat --isEncrypt=false --groupID=123456
go run crypto_example.go --default_redis_config="{\"address\":[\"127.0.0.1:6379\"]}" --key_manager_name=redis --inputFile=gx.dat --outputFile=x.dat --isEncrypt=false --isFinalOutput --groupID=1234567
*/

func getKvStorage() (kv.KV, error) {
	if gStorage == nil {
		gStorage, _ = kv.CreateKV(kv.StorageType(*kvType), &opt.Option{
			Address:  *kvAddress,
			Password: *kvPassword,
		})
	}
	return gStorage, nil

}

func main() {
	flagx.Parse()

	fi, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	br := bufio.NewReader(fi)
	storage, err := getKvStorage()
	if err != nil {
		fmt.Println("CreateSessionKVFailed")
	}

	defer fi.Close()

	for {

		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		line := string(a)
		fmt.Println(line)
		res, err := storage.Get(line)
		if err != nil {
			fmt.Println("Get error", err)
		}
		fmt.Printf("result:%s\n", res)

	}

	// go http.ListenAndServe("0.0.0.0:2333", nil)

	// if err := crypto.InitCrypto(); err != nil {
	// 	panic(err)
	// }
	// var appidList []string
	// api, _ := l5.NewDefaultApi()

	// appidList = append(appidList, "wx0cd0c0a6b01fb818")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx0cd0c0a6b01fb818")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx0cd0c0a6b01fb818")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx0cd0c0a6b01fb818")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")
	// appidList = append(appidList, "wx6173eb6085accfec")

	// var wg sync.WaitGroup
	// wuidHandler := dataprocess.NewWuidHandler("/data/xiaoyangma/part-r-00000", "111", appidList, api, &wg)
	// wuidHandler.Run()

	// wg.Wait()
	// fmt.Println("Done")
	// var encData string
	// var decData string
	// encData, _ = crypto.Encrypt("12345", "abc")
	// encData, _ = crypto.Encrypt("123456", encData)

	// decData, _ = crypto.Decrypt("12345", encData)
	// decData, _ = crypto.Decrypt("123456", decData)

	// fmt.Println(decData)
	// fmt.Println(crypto.Km.GetDecryptKey("12345"))
	// fmt.Println(crypto.Km.GetEncryptKey("12345"))

	// api, _ := l5.NewDefaultApi()
	// srv, _ := api.GetServerBySid(64948993, 65536)
	// fmt.Println(srv.Ip().String(), srv.Port())

	// fi, err := os.Open(*inputFile)
	// if err != nil {
	// 	fmt.Printf("Error: %s\n", err)
	// }
	// defer fi.Close()

	// fo, err := os.Create(*outputFile)
	// if err != nil {
	// 	fmt.Printf("Error: %s\n", err)
	// }
	// defer fo.Close()

	// br := bufio.NewReader(fi)
	// w := bufio.NewWriter(fo)
	// for {
	// 	a, _, c := br.ReadLine()
	// 	if c == io.EOF {
	// 		break
	// 	}
	// 	line := string(a)
	// 	if *isFirstInput {
	// 		line = util.StringToHex(line)
	// 	}
	// 	var encData string
	// 	if *isEncrypt {
	// 		encData, err = crypto.Encrypt(*groupID, line)
	// 	} else {
	// 		encData, err = crypto.Decrypt(*groupID, line)
	// 	}

	// 	if *isFinalOutput {
	// 		encData = util.HexToString(encData)
	// 	}

	// 	if err != nil {
	// 		glog.Errorf("failed to encrypt")
	// 		return
	// 	}
	// 	fmt.Fprintln(w, encData)
	// }
	// w.Flush()
}
