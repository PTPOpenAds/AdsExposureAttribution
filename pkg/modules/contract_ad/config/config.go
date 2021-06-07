/*
 * @Author: xiaoyangma@tencent.com
 * @Date: 2021-02-16 10:56:45
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-05-14 12:17:51
 */

package config

import (
	"github.com/golang/glog"
	"github.com/spf13/viper"
)

// Configuration :全局Config
var Configuration Config

type WxOpenid struct {
	OpenidService string `yaml:"openidService"`
	L5ModId       int    `yaml:"l5ModId"`
	L5CmdId       int    `yaml:"l5CmdId"`
	Bid           string `yaml:"bid"`
	Token         string `yaml:"token"`
}

// Config : config
type Config struct {
	WxOpenid    WxOpenid          `yaml:"WxOpenid"`
	ServiceConf map[string]string `yaml:"ServiceConf"`
	CoreDb      map[string]string `yaml:"CoreDb"`
	MrConf      map[string]string `yaml:"MrConf"`
	IsTest      bool              `yaml:"IsTest"`
	DB          map[string]string `yaml:"DB"`
	SQL         map[string]string `yaml:"SQL"`
}

func init() {
	viper.SetConfigName("contract_ad")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./conf")

	if err := viper.ReadInConfig(); err != nil {
		glog.Errorf("read config failed! err:%s", err)
	}

	viper.Unmarshal(&Configuration)
}
