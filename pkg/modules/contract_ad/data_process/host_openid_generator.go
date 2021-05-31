/*
 * @Author: xiaoyangma@tencent.com
 * @Date: 2021-01-23 21:28:54
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-05-14 12:20:57
 */

package dataprocess

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"sync"

	"git.code.oa.com/components/l5"
	"github.com/TencentAd/attribution/attribution/pkg/common/define"
	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/config"
	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/utility"
	"github.com/golang/glog"
)

var (
	OpenidSuffix  = "_openid"
	openidService = flag.String("openid_service", config.Configuration.WxOpenid.OpenidService, "openid_service")
	modID         = flag.Int("mod_id", config.Configuration.WxOpenid.L5ModId, "mod_id")
	cmdID         = flag.Int("cmd_id", config.Configuration.WxOpenid.L5CmdId, "cmd_id")
	bid           = flag.String("bid", config.Configuration.WxOpenid.Bid, "bid")
	token         = flag.String("token", config.Configuration.WxOpenid.Token, "token")
)

type WuidHandler struct {
	filepath      string
	attributionID string
	appidList     []string
	l5Api         l5.Api
	ch            chan string
	doneWG        *sync.WaitGroup
}

type SaveOpenid struct {
	Wuid   string `json:"wuid"`
	Appid  string `json:"appid"`
	Openid string `json:"openid"`
}

type wuidResponse struct {
	Data    retOpenid `json:"data"`
	Message string    `json:"msg"`
	RetCode int       `json:"ret"`
}

type retOpenid struct {
	Openid string `json:"openid"`
}

func NewWuidHandler(filePath string, attributionID string, appidList []string, l5Api *l5.Api, doneWG *sync.WaitGroup) *WuidHandler {
	wuidHandle := &WuidHandler{
		filepath:      filePath,
		attributionID: attributionID,
		appidList:     appidList,
		l5Api:         *l5Api,
		ch:            make(chan string),
		doneWG:        doneWG,
	}
	wuidHandle.doneWG.Add(1)
	go func() {
		defer wuidHandle.doneWG.Done()
		defer utility.TouchFile(wuidHandle.filepath + OpenidSuffix + utility.DefaultSuffix)
		utility.WriteToFile(wuidHandle.ch, wuidHandle.filepath+OpenidSuffix)
	}()
	return wuidHandle
}

// Run : 加载文件数据  这里只允许是单个文件
func (w *WuidHandler) Run() error {
	glog.V(define.VLogLevel).Info("WuidHandler.Run", w.filepath)
	lp := utility.NewLineProcess(w.filepath, w.processLine, func(line string, err error) {
		glog.Errorf("failed to handle line[%s], err[%v]", line, err)
	})

	if err := lp.RunNLines(*utility.DefaultNlines, utility.DefaultSeparator); err != nil {
		return err
	}
	lp.WaitDone()

	close(w.ch)
	return nil
}

func (w *WuidHandler) processLine(lines string) error {
	lineList := strings.Split(lines, utility.DefaultSeparator)
	srv, _ := w.l5Api.GetServerBySid(int32(*modID), int32(*cmdID))
	urlPath := fmt.Sprintf("http://%s:%d%s", srv.Ip().String(), srv.Port(), *openidService)
	paramsMap := make(map[string]string)

	var wuidResp wuidResponse
	var saveData SaveOpenid
	var stringData []byte
	for _, wuid := range lineList {
		for _, appID := range w.appidList {
			paramsMap["openid"] = wuid
			paramsMap["appid"] = appID
			paramsMap["bid"] = *bid
			paramsMap["token"] = *token

			err := utility.DoHTTPGet(urlPath, utility.ContentType, &paramsMap, &wuidResp)
			if err != nil {
				return err
			}
			saveData = SaveOpenid{
				Openid: wuidResp.Data.Openid,
				Wuid:   wuid,
				Appid:  appID,
			}
			stringData, _ = json.Marshal(saveData)
			w.ch <- string(stringData)
		}
	}

	return nil
}
