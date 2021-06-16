/*
 * @Author: xiaoyangma@tencent.com
 * @Date: 2021-02-03 21:14:49
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-03-03 14:59:45
 */

package utility

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/TencentAd/attribution/attribution/pkg/common/define"
	"github.com/TencentAd/attribution/attribution/pkg/modules/contract_ad/config"
	"github.com/golang/glog"
)

var (
	mrRetry = flag.Int("mr_retry_times", 5, "mr_retry_times")
)

var (
	mrConf = config.Configuration.MrConf
)

// ScriptHandler ScriptHandler
type ScriptHandler struct {
	JobID             string
	Dates             string
	ConvDates         string
	CompetingDates    string
	Mobile            string
	Theta             string
	EmbeddingFilePath string
	ConfFilePath      string
}

// RunShell 调用shell脚本并同步等待执行
func RunShell(shellCommandPath string, args string) error {
	command := fmt.Sprintf("sh %s %s", shellCommandPath, args)
	return RunCommand(command)
}

// RunCommand : RunCommand
func RunCommand(command string) error {
	errInfo := fmt.Sprintf("RunCommand:(%s)", command)
	cmd := exec.Command("/bin/bash", "-c", command)
	if out, err := cmd.CombinedOutput(); err != nil {
		glog.Errorf("[ERROR] shell command failed! \n[Command] %s, \n[ERROR] %s\n[OUTPUT] %s",
			command, err, out)
	} else {
		glog.Infof("Output:%s\n%s", errInfo, out)
	}
	return nil
}

// RunCommandWithRetry
func RunCommandWithRetry(command string) error {
	var err error
	for i := 0; i < *mrRetry; i++ {
		err = RunCommandRealTime(command)
		if err == nil {
			break
		} else {
			time.Sleep(3 * time.Minute)
			glog.Infof("[INFO] Retry command %s", command)
		}
	}
	return err
}

// RunCommandRealTime : RunCommandRealTime
func RunCommandRealTime(command string) error {
	if config.Configuration.IsTest && strings.Contains(command, "hadoop") {
		command = "echo " + command
	}
	errInfo := fmt.Sprintf("RunCommandRealTime:(%s)", command)
	glog.V(define.VLogLevel).Infof(errInfo)
	cmd := exec.Command("/bin/bash", "-c", command)
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout // 命令的错误输出和标准输出都连接到同一个管道
	if err != nil {
		glog.Errorf("[ERROR] %s failed : %s", errInfo, err)
		return err
	}
	if err = cmd.Start(); err != nil {
		glog.Errorf("[ERROR] %s start failed : %s", errInfo, err)
		return err
	}
	// 从管道中实时获取输出并打印到终端
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		// remove null character (hexdump -C file)
		glog.Info(strings.Replace(string(bytes.Trim(tmp, "\x00")), "\n", "", -1))
		if err != nil {
			break
		}
	}
	if err = cmd.Wait(); err != nil {
		glog.Errorf("[ERROR] %s end failed : %s", errInfo, err)
		return err
	}
	return nil
}

// DumpData : DumpData
func DumpData(attributionID string, src string) string {
	dst := "data/imp/" + attributionID
	if !config.Configuration.IsTest {
		os.RemoveAll(dst)
	}
	return fmt.Sprintf("hadoop fs -copyToLocal %s %s", src, dst)
}

// LoadExpData : LoadExpData
func LoadExpData(dates string, adgroupIDs string, attributionID string) string {
	output := mrConf["uniquser"] + attributionID
	return fmt.Sprintf("hadoop jar %s %s %s %s %s %s %s %s %s", mrConf["jar"], mrConf["class"],
		mrConf["ugi"], mrConf["queue"], mrConf["mapred"], mrConf["impdata"], output,
		adgroupIDs, dates)
}
