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
    "context"
    "os/signal"
    "syscall"
    "time"

    "aliyun-dns/module/myip"

    myConfig "aliyun-dns/config"
    "aliyun-dns/module/aliddns"
    "aliyun-dns/module/help"
    "aliyun-dns/module/loger"
    alidns "github.com/alibabacloud-go/alidns-20150109/client"
    "github.com/alibabacloud-go/tea/tea"
)

func main() {
    if help.ParseArgs() {
        return
    }
    // 读取配置
    var cfg myConfig.Config
    if !cfg.LoadConfig(help.Cfg) {
        loger.PreInfoHeav("配置文件加载出错！")
        return
    }

    // 日志功能
    loger.InitLogger(cfg.Log)

    // 创建客户端
    cliPtr, err := aliddns.CreateClient(tea.String(cfg.AccessKeyId), tea.String(cfg.AccessKeySecret))
    if err != nil {
        loger.Error(err.Error())
        return
    }
    cli := *cliPtr

    // 加载缓存
    myip.CurrentCache.Init(&cfg, cli)

    // 根据宽带多拨情况来判断使用哪个函数主体运行
    var runFunc func(context.Context, *myConfig.Config, alidns.Client)
    if cfg.BroadbandRetry < 2 {
        loger.Info("正常DDNS")
        runFunc = aliddns.Run
    } else {
        loger.Info("宽带多拨DDNS")
        runFunc = aliddns.RunOnMultiBroadband
    }
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer func() {
        loger.Info("结束运行")
        stop()
    }()
    ctxChild, _ := context.WithCancel(ctx)

    // 用来整点运行
    interval := cfg.Interval * 1000
    t := time.NewTimer(next(interval))
    for {
        select {
        case <-ctx.Done():
            return
        case <-t.C:
            runFunc(ctxChild, &cfg, cli)
            t.Reset(next(interval))
        }
    }
}

// 用这种方式误差在10ms左右
// 下次触发定时器的时刻
func next(interval int64) time.Duration {
    return time.Duration(interval-(time.Now().UnixMilli()%interval)) * time.Millisecond
}
