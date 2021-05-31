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

	"github.com/TencentAd/attribution/attribution/pkg/common/flagx"
	"github.com/TencentAd/attribution/attribution/pkg/impression/kv"
	"github.com/TencentAd/attribution/attribution/pkg/impression/kv/opt"
)

var (
	gStorage   kv.KV = nil
	kvAddress        = flag.String("kv_address", "./db", "")
	kvPassword       = flag.String("kv_password", "", "")
	kvType           = flag.String("kv_type", "HASH_LEVELDB", "")
)

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
	storage, err := getKvStorage()
	if err != nil {
		fmt.Println(fmt.Errorf("CreateSessionKVFailed"))
		return
	}
	storage.Set("wx0cd0c0a6b01fb810", "wx0cd0c0a6b01fb800")
	storage.Set("wx0cd0c0a6b01fb811", "wx0cd0c0a6b01fb811")
	storage.Set("wx0cd0c0a6b01fb812", "wx0cd0c0a6b01fb822")
	storage.Set("wx0cd0c0a6b01fb813", "wx0cd0c0a6b01fb833")
	storage.Set("wx0cd0c0a6b01fb814", "wx0cd0c0a6b01fb844")
	storage.Set("wx0cd0c0a6b01fb815", "wx0cd0c0a6b01fb855")
	storage.Set("wx0cd0c0a6b01fb816", "wx0cd0c0a6b01fb866")
	storage.Set("wx0cd0c0a6b01fb817", "wx0cd0c0a6b01fb877")
	storage.Set("wx0cd0c0a6b01fb818", "wx0cd0c0a6b01fb888")
	storage.Set("wx0cd0c0a6b01fb819", "wx0cd0c0a6b01fb899")
	storage.Set("wx0cd0c0a6b01fb81a", "wx0cd0c0a6b01fb81A")
	storage.Set("wx0cd0c0a6b01fb81b", "wx0cd0c0a6b01fb81B")
	storage.Set("wx0cd0c0a6b01fb81c", "wx0cd0c0a6b01fb81C")
	storage.Set("wx0cd0c0a6b01fb81d", "wx0cd0c0a6b01fb81D")
	storage.Set("wx0cd0c0a6b01fb81e", "wx0cd0c0a6b01fb81E")
	storage.Set("wx0cd0c0a6b01fb81f", "wx0cd0c0a6b01fb81F")

	fmt.Println(storage.Get("wx0cd0c0a6b01fb810"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb811"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb812"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb813"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb814"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb815"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb816"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb817"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb818"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb819"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb81a"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb81b"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb81c"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb81d"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb81e"))
	fmt.Println(storage.Get("wx0cd0c0a6b01fb81f"))
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
