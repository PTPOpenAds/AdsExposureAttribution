/*
 * @Author: joyesjiang@tencent.com
 * @Date: 2021-01-17 22:58:03
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-01-27 11:01:16
 */

package utility

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/TencentAd/attribution/attribution/pkg/common/define"
	"github.com/golang/glog"
)

var (
	outputFileDefault = flag.String("outputFile", "output_matched_opendids.dat", "")
)

// WriteToFile 输出指定Channel中的内容到文件中
func WriteToFile(ch chan string, outputFile string) bool {
	if outputFile == "" {
		outputFile = *outputFileDefault
	}
	var writer *bufio.Writer
	var lineNum = 0
	for {
		if saveConvData, ok := <-ch; ok {
			if lineNum == 0 {
				glog.V(define.VLogLevel).Info("Start Do writeToFile, filePath: ", outputFile)
				fo, err := os.Create(outputFile)
				if err != nil {
					glog.Error("Error fo Create : ", err.Error())
					return false
				}
				defer fo.Close()
				writer = bufio.NewWriter(fo)
			}
			lineNum++
			fmt.Fprintln(writer, saveConvData)
			glog.V(define.VLogLevel).Infof("writeToFile %v: %v ", outputFile, saveConvData)
		} else {
			glog.Info("Finish Do writeToFile: ", outputFile, ", lineNum: ", lineNum)
			break
		}
	}
	if writer != nil {
		writer.Flush()
	}
	return true
}

// Exists 判断path是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// IsDir 判断path是否是文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile 判断path是否是文件
func IsFile(path string) bool {
	return Exists(path) && !IsDir(path)
}

// GetFileList 获得路径下所有文件，非递归
func GetFileList(path string) []string {
	var files []string
	fs, _ := ioutil.ReadDir(path)
	for _, file := range fs {
		files = append(files, fmt.Sprintf("%s/%s", path, file.Name()))
	}
	return files
}

// TouchFile 生成一个空文件
func TouchFile(path string) error {
	if IsFile(path) || IsDir(path) {
		return nil
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}
