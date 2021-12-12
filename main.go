/*
Copyright 2021 qingchuwudi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"aliyun-dns/module/myip"
	"time"

	myConfig "aliyun-dns/config"
	"aliyun-dns/module/aliddns"
	"aliyun-dns/module/help"
	"aliyun-dns/module/loger"
	alidns "github.com/alibabacloud-go/alidns-20150109/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/antlabs/timer"
)

func main() {
	if help.ParseArgs() {
		return
	}
	// 读取配置
	var config myConfig.Config
	if !config.LoadConfig(help.Cfg) {
		loger.PreInfoHeav("配置文件加载出错！")
		return
	}

	// 日志功能
	loger.InitLogger(config.LogPath, "aliyun-dns.log")

	// 创建客户端
	cli, err := aliddns.CreateClient(tea.String(config.AccessKeyId), tea.String(config.AccessKeySecret))
	if err != nil {
		loger.Error(err.Error())
		return
	}

	// 加载缓存
	myip.CurrentCache.Init(&config, cli)

	// 根据宽带多拨情况来判断使用哪个函数主体运行
	var runFunc func(config2 *myConfig.Config, client *alidns.Client)
	if config.BroadbandRetry < 2 {
		loger.Info("正常DDNS")
		runFunc = aliddns.Run
	} else {
		loger.Info("宽带多拨DDNS")
		runFunc = aliddns.RunOnMultiBroadband
	}
	tm := timer.NewTimer()

	tm.ScheduleFunc(30*time.Second, func() {
		runFunc(&config, cli)
	})

	tm.Run()
}
