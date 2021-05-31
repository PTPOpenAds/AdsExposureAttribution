/*
 * copyright (c) 2021, Tencent Inc.
 * All rights reserved.
 *
 * Author:  joyesjiang@tencent.com
 * Create Time: 01/11/21
 */

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/TencentAd/attribution/attribution/pkg/common/flagx"
	"github.com/TencentAd/attribution/attribution/pkg/common/metricutil"
	"github.com/TencentAd/attribution/attribution/pkg/crypto"
	"github.com/TencentAd/attribution/attribution/pkg/handler/http/encrypt/safeguard"
	"github.com/TencentAd/attribution/attribution/pkg/impression/kv"
	"github.com/TencentAd/attribution/attribution/pkg/impression/kv/opt"
	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/config"
	dataprocess "github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/data_process"
	Define "github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/define"
	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/session"
	"github.com/golang/glog"
)

var (
	role                 = flag.String("role", "HOST", "")
	serverAddress        = flag.String("server_address", ":9081", "")
	metricsServerAddress = flag.String("metric_server_address", ":8005", "")

	// 向Host侧发送各类请求时对应的服务地址或者域名
	hostServerPushAddress = flag.String("host_server_address", "http://localhost:9081", "")
	// 向Guest侧发送各类请求时对应的服务地址或者域名
	guestServerPushAddress = flag.String("guest_server_address", "http://localhost:9080", "")

	/*********************************/
	// 已下为Host模块负责监听的服务对应的路径：
	// Guest --> Host 发送会话创建的请求对应的服务路径
	hostCreateSessionSvrPath  = flag.String("create_session_pattern", "/host/create_session", "")
	guestCreateSessionSvrPath = flag.String("create_guest_session_pattern", "/guest/create_session", "")
	// shell --> Host 触发Host侧开始处理Imp数据并开始和Guest侧的二次加密交互
	hostImpProcessSvrPath = flag.String("imp_process_pattern", "/host/imp_process", "")
	// Guest --> Host 发送g(openid_conv)到Host进行conv匹配对应的服务路径
	hostConvMatchSvrPath = flag.String("host_conv_match_svr_pattern", "/host/conv_data_match", "")
	// Guest --> Host 推送最终匹配的conv对应的openid明文
	hostSaveMatchedConvSvrPath = flag.String("host_save_matched_listen_pattern", "/host/save_matched_conv_data", "")
	// Host侧用于存储匹配结果的路径
	hostMatchResultSavePath = flag.String("host_match_result_path", "data/host_match_result", "")

	/*********************************/
	// 以下为Guest模块负责监听的服务对应的路径：
	// Host --> Guest 发送f(openId_imp)到guest进行二次加密 --- Guest侧Listen目录
	guestEncryptImpSvrPath = flag.String("guest_encrypt_svr_pattern", "/guest/crypto_encrypt", "")
	// Host --> Guest 触发Guest侧开始处理Conv数据
	guestConvProcessSvrPath = flag.String("guest_conv_process_svr_pattern", "/guest/conv_process", "")
	// Guest侧用于存储匹配结果的路径
	guestMatchResultSavePath = flag.String("guest_match_result_path", "data/guest_match_result/", "")

	// 以下为Guest初始化一次归因需要填写的参数
	// accountID   = flag.Int("account_id", 12345, "account_id")
	// campaignIDs = flag.String("campaign_ids", "abcdef,qwert", "campaign_ids")
	// appIDs      = flag.String("app_ids", "app_id1,app_id2", "app_ids")

	// 推荐使用redis, levelDB 查询速度影响性能
	kvType           = flag.String("kv_type", config.Configuration.ServiceConf["kvtype"], "")
	kvAddress        = flag.String("kv_address", config.Configuration.ServiceConf["kvaddress"], "")
	kvPassword       = flag.String("kv_password", config.Configuration.ServiceConf["kvpassword"], "")
	gStorage   kv.KV = nil
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

// Guest --> Host 发送会话创建的请求
func hostCreateSessionHandle() (bool, error) {
	storage, err := getKvStorage()
	if err != nil {
		return false, fmt.Errorf("CreateSessionKVFailed")
	}

	http.Handle(*hostCreateSessionSvrPath, session.NewSessionHandler(storage))
	glog.Info("do hostCreateSessionHandle succ: ", *hostCreateSessionSvrPath)
	return true, err
}

func guestCreateSessionHandle() (bool, error) {
	storage, err := getKvStorage()
	if err != nil {
		return false, fmt.Errorf("CreateSessionKVFailed")
	}
	http.Handle(*guestCreateSessionSvrPath, session.NewGuestSessionHandler(storage, *hostServerPushAddress+*hostCreateSessionSvrPath))
	glog.Info("do hostCreateSessionHandle succ: ", *guestCreateSessionSvrPath)
	return true, err
}

// Host对应曝光数据进行处理
func hostImpProcessHandle() (bool, error) {
	storage, err := getKvStorage()
	if err != nil {
		glog.Infof("%s\n", err)
	}

	glog.Info("do hostImpProcessHandle storage: ", storage)
	guestEncryptImpPushSvr := *guestServerPushAddress + *guestEncryptImpSvrPath
	guestConvProcessPushSvr := *guestServerPushAddress + *guestConvProcessSvrPath
	http.Handle(*hostImpProcessSvrPath, dataprocess.NewImpressionFileHandle(guestEncryptImpPushSvr, guestConvProcessPushSvr, storage))
	glog.Info("do hostImpProcessHandle succ: ", *hostImpProcessSvrPath)
	return true, err
}

// Host侧执行Conv和Imp的Match过程
func hostDoConvMatchHandle() (bool, error) {
	storage, err := getKvStorage()
	if err != nil {
		glog.Infof("%s\n", err)
	}

	glog.Info("do hostDoConvMatchHandle storage: ", storage)
	http.Handle(*hostConvMatchSvrPath, dataprocess.NewHostConvMatchHandler(storage))
	glog.Info("do hostDoConvMatchHandle succ: ", hostConvMatchSvrPath)
	return true, nil
}

// Host侧保存Match到的conv对应的openId原始数据
func hostSaveMatchedHandle() (bool, error) {
	http.Handle(*hostSaveMatchedConvSvrPath, dataprocess.NewHostSaveMatchResultHandler(*hostMatchResultSavePath))
	glog.Info("do hostSaveMatchedHandle succ: ", *hostSaveMatchedConvSvrPath)
	return true, nil
}

// Guest接收Host侧的曝光加密信息f(openid)并进行加密
func guestImpCryptoHandle() (bool, error) {
	encryptSafeguard, err := safeguard.NewConvEncryptSafeguard()
	if err != nil {
		glog.Errorf("failed to create encrypt safeguard, err: %v", err)
		return false, err
	}
	http.Handle(*guestEncryptImpSvrPath, dataprocess.NewCryptoHandle().WithSafeguard(encryptSafeguard))
	glog.Info("do guestImpCryptoHandle succ: ", *guestEncryptImpSvrPath)
	return true, err
}

// Host --> Guest 触发Guest侧开始处理Conv数据并开始和Host侧进行Conv匹配检测交互，并返回匹配的openid明文
func guestConvProcessHandle() (bool, error) {
	storage, err := getKvStorage()
	if err != nil {
		glog.Infof("%s\n", err)
	}
	hostConvMatchPushSvr := *hostServerPushAddress + *hostConvMatchSvrPath
	hostSaveMatchedPushSvr := *hostServerPushAddress + *hostSaveMatchedConvSvrPath
	http.Handle(*guestConvProcessSvrPath, dataprocess.NewConvProcessFileHandle(
		*guestMatchResultSavePath, hostConvMatchPushSvr, hostSaveMatchedPushSvr, storage))
	glog.Info("do guestConvProcessHandle succ: ", *guestConvProcessSvrPath)
	return true, nil
}

// func guestInitSession() (bool, error) {
// 	paramsMap := make(map[string]string)
// 	paramsMap["account_id"] = strconv.Itoa(*accountID)
// 	paramsMap["campaign_ids"] = *campaignIDs
// 	paramsMap["app_ids"] = *appIDs
// 	paramsMap["trace_id"] = strconv.FormatInt(time.Now().UnixNano(), 10)
// 	var sessResp session.CreateSessionResponse

// 	err := utility.DoHTTPGet(*hostServerPushAddress+*hostCreateSessionSvrPath, utility.ContentType, &paramsMap, &sessResp)
// 	if err != nil {
// 		glog.Info("Do http get failed! :", err)
// 		return false, err
// 	}

// 	crypto.Km.GetEncryptKey(sessResp.AttributionID)
// 	glog.Infof("Create Session Response: %+v\n", sessResp)

// 	return true, nil
// }

// serveHttp 主服务
// TODO：前期(线下文件交互)实现为：接受请求后开始加载指定的文件并且执行加解密操作
// 后续规划：通过Http请求实现大文件交互，并直接基于大文件内的数据进行加解密操作
func serveHTTP() error {
	if err := crypto.InitCrypto(); err != nil {
		panic(err)
	}

	switch *role {
	case Define.RoleHost:
		if ok, err := hostCreateSessionHandle(); !ok {
			glog.Errorf("do hostCreateSessionHandle Failed, abort !")
			return err
		}

		if ok, err := hostImpProcessHandle(); !ok {
			glog.Errorf("do hostImpProcessHandle Failed, abort !")
			return err
		}

		if ok, err := hostDoConvMatchHandle(); !ok {
			glog.Errorf("do hostDoConvMatchHandle Failed, abort !")
			return err
		}

		if ok, err := hostSaveMatchedHandle(); !ok {
			glog.Errorf("do hostSaveMatchedHandle Failed, abort !")
			return err
		}

	case Define.RoleGuest:
		if ok, err := guestCreateSessionHandle(); !ok {
			glog.Errorf("do guestCreateSessionHandle Failed, abort !")
			return err
		}
		if ok, err := guestImpCryptoHandle(); !ok {
			glog.Errorf("do guestImpCryptoHandle Failed, abort !")
			return err
		}

		if ok, err := guestConvProcessHandle(); !ok {
			glog.Errorf("do guestConvProcessHandle Failed, abort !")
			return err
		}
		// if ok, err := guestInitSession(); !ok {
		// 	glog.Errorf("do guestInitSession Failed, abort !")
		// 	return err
		// }
	default:
		glog.Errorf("do serveHTTP parse role Failed: %v", role)
		return fmt.Errorf("do serveHTTP parse role Failed: %v", role)
	}

	glog.Info("server init done, going to start")
	return http.ListenAndServe(*serverAddress, nil)
}

func main() {
	if err := flagx.Parse(); err != nil {
		panic(err)
	}
	_ = metricutil.ServeMetrics(*metricsServerAddress)
	if err := serveHTTP(); err != nil {
		glog.Errorf("failed to start server, err: %v", err)
		os.Exit(1)
	}
}
