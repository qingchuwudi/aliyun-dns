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
package config

import (
	"os"

	"github.com/goccy/go-yaml"

	"aliyun-dns/module/filecheck"
	"aliyun-dns/module/loger"
)

const (
	IPv4Type  = "A"
	IPv6Type  = "AAAA"
	AliyunDNS = "dns.aliyuncs.com" // 修改DNS的API地址
)

// ------------------------------------------------------------------------
//
// ------------------------------------------------------------------------

// 配置文件
type Config struct {
	AccessKeyId     string           `yaml:"accessKeyId"`
	AccessKeySecret string           `yaml:"accessKeySecret"`
	Log             *loger.LogConfig `yaml:"log"`
	TTL             int64            `yaml:"ttl"`
	IPv4            string           `yaml:"ipv4_check_url"`
	IPv6            string           `yaml:"ipv6_check_url"`
	Interval        int64            `yaml:"interval"`
	BroadbandRetry  int8             `yaml:"broadband_retry"`
	UseCache        bool             `yaml:"cache"`
	Customer        []Customer       `yaml:"customer"`
	// IP        map[string]string // 记录阿里云dns的ip
	// RecordId       map[string]string // 记录阿里云DNS的recordID
}

type Customer struct {
	Domain string `yaml:"domain"`
	IPv4RR string `yaml:"ipv4_rr"`
	IPv6RR string `yaml:"ipv6_rr"`
}

// 从配置文件加载配置
func (c *Config) LoadConfig(file string) (success bool) {
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		loger.PreError(err.Error())
		return false
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		loger.PreError(err.Error())
		return false
	}
	// log
	if c.Log == nil {
		c.Log = &loger.LogConfig{
			Path:    "",
			Level:   "debug",
			Develop: false,
		}
	}
	if !IsLogValid(c.Log) {
		return false
	}
	// 检查 ip
	if (c.IPv4 == "") && (c.IPv6 == "") {
		loger.PreError("'ipv4_check_url' 和 'ipv6_check_url' 至少配置一个！")
		return false
	}

	// 周期最短5秒
	if c.Interval < 5 {
		c.Interval = 5
	}
	return true
}

// 初始化
func (c *Config) InitBroadbandRecords() map[string]bool {
	if c.BroadbandRetry > 0 {
		return make(map[string]bool, c.BroadbandRetry)
	}
	return nil
}

// 检查日志配置有效性
func IsLogValid(l *loger.LogConfig) bool {
	if l.Path != "" && !filecheck.IsDir(l.Path) {
		loger.PreError("日志路径(%s)配置有误：路径不存在或没有权限！", l.Path)
		return false
	}
	switch l.Level {
	case "debug", "info", "warn", "error":
		break
	default:
		loger.PreError("日志等级(%s)配置有误", l.Level)
		return false
	}
	return true
}
