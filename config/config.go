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
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	//"github.com/alibabacloud-go/tea/tea"
	"gopkg.in/yaml.v2"
)

const (
    IPv4Type = "A"
    IPv6Type = "AAAA"
    AliyunDNS = "dns.aliyuncs.com" // 修改DNS的API地址
)

// 配置文件
type Config struct {
	AccessKeyId     string     `yaml:"accessKeyId"`
	AccessKeySecret string     `yaml:"accessKeySecret"`
	TTL             int64      `yaml:"ttl"`
	IPv4            string     `yaml:"ipv4_check_url"`
	IPv6            string     `yaml:"ipv6_check_url"`
	Interval        int        `yaml:"interval"`
	Customer        []Customer `yaml:"customer"`
}

type Customer struct {
	Domain string `yaml:"domain"`
	IPv4RR string `yaml:"ipv4_rr"`
	IPv6RR string `yaml:"ipv6_rr"`
}

func (c *Config) InitConfig(file string) error {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return err
	}
	// 检查
	if (c.IPv4 == "") && (c.IPv6 == "") {
		return errors.New("It has at least one value of 'ipv4_check_url' and 'ipv6_check_url' in config.yaml.")
	}
	return nil
}

func (c *Config) PublilcIPs() (PubIPv4, PubIPv6 *string) {
	PubIPv4, PubIPv6 = nil, nil
	if c.IPv4 != "" {
		PubIPv4 = GetPublishIP(c.IPv4)
	}
	if c.IPv6 != "" {
		PubIPv6 = GetPublishIP(c.IPv6)
	}
	return PubIPv4, PubIPv6
}


func GetPublishIP(IPCheckUrl string) *string {
	resp, err := http.Get("http://" + IPCheckUrl)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	ip1 := strings.Replace(string(content), "\n", "", -1)
	ip2 := strings.Replace(ip1, "\r", "", -1)
	return &ip2
}

