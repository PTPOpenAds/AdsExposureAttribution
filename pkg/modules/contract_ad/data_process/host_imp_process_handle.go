/*
 * copyright (c) 2019, Tencent Inc.
 * All rights reserved.
 *
 * Author:  linceyou@tencent.com
 * Last Modify: 8/12/20, 5:35 PM
 */

package dataprocess

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"git.code.oa.com/components/l5"
	"github.com/TencentAd/attribution/attribution/pkg/common/define"
	"github.com/TencentAd/attribution/attribution/pkg/crypto"
	"github.com/TencentAd/attribution/attribution/pkg/crypto/util"
	"github.com/TencentAd/attribution/attribution/pkg/impression/kv"
	dataprotocal "github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/protocal"
	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/utility"

	"github.com/golang/glog"
)

var (
	defaultImpPath = flag.String("default_imp_path", "data/imp/", "default_imp_path")
)

// FileHandle : 定义处理曝光handler
type FileHandle struct {
	kv                       kv.KV
	filepath                 string
	attributionID            string
	guestEncryptImpPushPath  string
	guestConvProcessPushPath string
}

// Data : 定义输入impression
type Data struct {
	Wuid      string `json:"wuid"`
	Openid    string `json:"openid"`
	Timestamp int    `json:"timestamp"`
}

// 用于从http.request中解析对应的字段
const (
	attributionIDKey = "attribution_id"
	filepathKey      = "file_path"
	appidKey         = "app_ids"

	noAttributionIDKeyError = "can't find attribution_id"
	noFilepathKeyError      = "can't find file_path"
	sendSuffix              = "_send"
)

type impFileProcessRequest struct {
	AttributionID string   `json:"123456"`         // 标记本次的归因的任务ID(accountid_amsid)
	Filepath      string   `json:"/data/imp_file"` // 需要归因的排期ID列表
	AppidList     []string `json:"appids"`
}

type ImpFileProcessResponse struct {
	Retcode       int64  `json:"ret_code"`       // 当前请求处理的状态
	Message       string `json:"message"`        // 消息内容
	AttributionID string `json:"attribution_id"` // 标记本次的归因的任务ID(accountid_amsid)
}

func (p *ImpFileProcessResponse) Init(attributionID string) {
	p.Retcode = 200
	p.Message = "Succ"
	p.AttributionID = attributionID
	glog.Infof("impFileProcessResponse Init: %v ", *p)
}

func (p *ImpFileProcessResponse) SetFailedResponse(errMsg string) {
	p.Retcode = 500
	p.Message = errMsg
	glog.Infof("impFileProcessResponse SetFailedResponse: %v ", *p)
}

func ConstructKey(attributionID string, openid string) string {
	return "attributionID:" + attributionID + ",impGOpenID:" + openid
}

// ImpProcessHandler 用于处理曝光文件数据的Handler
type ImpProcessHandler struct {
	fileHandle *FileHandle
	l5Api      *l5.Api
}

// NewImpressionFileHandle : 构造FileHandle
func NewImpressionFileHandle(guestEncryptImpPushSvr string, guestConvProcessPushSvr string, kv kv.KV) *ImpProcessHandler {
	glog.Info("NewImpressionFileHandle push guest_erypt path: ", guestEncryptImpPushSvr)
	glog.Info("NewImpressionFileHandle push guest_conv path: ", guestConvProcessPushSvr)
	l5Api, _ := l5.NewDefaultApi()
	return &ImpProcessHandler{
		fileHandle: &FileHandle{
			kv:                       kv,
			guestEncryptImpPushPath:  guestEncryptImpPushSvr,
			guestConvProcessPushPath: guestConvProcessPushSvr,
		},
		l5Api: l5Api,
	}
}

func (handle *ImpProcessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	errHandle := func(err error) {
		log.Print(err)
		_, _ = w.Write([]byte(err.Error()))
	}

	values := r.URL.Query()
	glog.Infof("Display_ImpProcessHandler req: %v", values)

	var impReq impFileProcessRequest
	var impResp ImpFileProcessResponse

	impReq.AttributionID = string(values.Get(attributionIDKey))
	if impReq.AttributionID == "" {
		errHandle(fmt.Errorf(noAttributionIDKeyError))
		impResp.SetFailedResponse(noAttributionIDKeyError)
		return
	}
	impResp.Init(impReq.AttributionID)

	impReq.Filepath = *defaultImpPath + impReq.AttributionID
	impReq.AppidList = strings.Split(string(values.Get(appidKey)), ",")

	//handle.fileHandle.filepath = impReq.Filepath
	handle.fileHandle.attributionID = impReq.AttributionID
	glog.Info("Ready_to_run, imp_file_path: ", impReq.Filepath)

	// 将状态以及进度写入db
	utility.UpdateJobStatus(impReq.AttributionID, "imp", 0)

	defer func() {
		glog.Infof("StartRunFileHandle: %s", impReq.Filepath)
		fileList := utility.GetFileList(impReq.Filepath)

		// 换算每个文件对应的进度
		totalFileNum := 0
		for _, filePath := range fileList {
			filePathList := strings.Split(filePath, "/")
			if strings.Contains(filePathList[len(filePathList)-1], "_") {
				continue
			}
			totalFileNum += 1
		}
		singleFileProcessRate := 100.0 / totalFileNum
		fileProcessed := 0

		var wuidWG sync.WaitGroup
		var impWG sync.WaitGroup

		fileChannel := make(chan string, 100)

		impWG.Add(1)
		go func() {
			defer impWG.Done()
			for {
				if filePath, ok := <-fileChannel; ok {
					if !utility.Exists(filePath + sendSuffix) {
						handle.fileHandle.Run(filePath)
						utility.TouchFile(filePath + sendSuffix)
						glog.Infof("File %s has already been Encrypted!", filePath)
					}
					//加密曝光不会并行，可以直接进行累加，然后更新状态
					fileProcessed += 1
					utility.UpdateJobStatus(impReq.AttributionID, "imp", int(singleFileProcessRate*fileProcessed))
				} else {
					glog.Info("All imps has been encrypted!")
					utility.UpdateJobStatus(impReq.AttributionID, "imp", 100)
					break
				}
			}
		}()

		for _, filePath := range fileList {
			filePathList := strings.Split(filePath, "/")
			if strings.Contains(filePathList[len(filePathList)-1], "_") {
				continue
			}

			if !utility.Exists(filePath + OpenidSuffix + utility.DefaultSuffix) {
				wuidHandle := NewWuidHandler(filePath, handle.fileHandle.attributionID, impReq.AppidList, handle.l5Api, &wuidWG)
				wuidHandle.Run()
				wuidWG.Wait()
			}
			glog.Infof("File %s has already convertered to openid!", filePath)
			fileChannel <- filePath + OpenidSuffix
		}
		close(fileChannel)

		impWG.Wait()
		handle.fileHandle.doGetConvRequest(impReq.AttributionID)
	}()

	data, err := json.Marshal(impResp)
	glog.Infof("encrypt response: %s", string(data))

	if err != nil {
		w.WriteHeader(500)
	} else {
		_, _ = w.Write(data)
	}
}

// Run : 加载文件数据
func (p *FileHandle) Run(fileName string) error {
	glog.Info("FileHandle.Run", fileName)
	lp := utility.NewLineProcess(fileName, p.processLine, func(line string, err error) {
		glog.Errorf("failed to handle line[%s], err[%v]", line, err)
	})

	if err := lp.RunNLines(*utility.DefaultNlines, utility.DefaultSeparator); err != nil {
		return err
	}
	lp.WaitDone()
	return nil
}

func (p *FileHandle) processLine(lines string) error {
	lineList := strings.Split(lines, utility.DefaultSeparator)
	var saveData SaveOpenid
	// 加密数据构建RequestData
	resq := &dataprotocal.CryptoRequest{
		AttributionID: p.attributionID,
		DataType:      dataprotocal.Imp,
	}
	for _, line := range lineList {
		if err := json.Unmarshal([]byte(line), &saveData); err != nil {
			glog.Errorf("Failed to parse impression line: %s, error %s.", line, err)
		}
		glog.V(define.VLogLevel).Infof("processLine: %+v", saveData)

		ImpFOpenid, err := crypto.Encrypt(p.attributionID, util.StringToHex(saveData.Openid))
		if err != nil {
			return err
		}
		resqData := &dataprotocal.RequestData{
			Openid: ImpFOpenid,
		}
		glog.V(define.VLogLevel).Info("processLine ImpFOpenid: ", ImpFOpenid)
		resq.Data = append(resq.Data, resqData)

	}

	var resp dataprotocal.CryptoResponse
	if err := utility.DoHTTPPost(p.guestEncryptImpPushPath, utility.ContentType, resq, &resp); err != nil {
		glog.Info("Send Imp data failed: ", err)
	}

	if len(resp.Data) != len(lineList) {
		glog.Info("Impression some encrypted data lost")
	}

	for _, respData := range resp.Data {
		ImpGOpenid, _ := crypto.Decrypt(p.attributionID, respData.Openid)

		// 增加attributinID 作为key，保证多个归因任务可以并行
		err := p.kv.Set(ConstructKey(p.attributionID, ImpGOpenid), "imp")
		if err != nil {
			glog.Info("Set Kv error: ", err)
		}
	}
	return nil
}

func (p *FileHandle) doGetConvRequest(attributionID string) error {
	glog.Info("FileHandle DoGetConvRequest: ", attributionID)
	resqData := &dataprotocal.GetConvDataRequest{
		AttributionID: attributionID,
		DataType:      dataprotocal.Imp,
	}
	var resp dataprotocal.GetConvDataResponse
	utility.DoHTTPPost(p.guestConvProcessPushPath, utility.ContentType, resqData, &resp)
	utility.UpdateJobStatus(attributionID, "convMatch", 0)
	// 打印请求和返回结果
	glog.Infof("processLine doGetConvRequest, request: %v, resp: %v", resqData, resp)
	return nil
}
