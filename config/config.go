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
	"fmt"
	alidns "github.com/alibabacloud-go/alidns-20150109/client"
	"io/ioutil"
	"net/http"
	"strings"
	"gopkg.in/yaml.v2"
)

const (
	IPv4Type  = "A"
	IPv6Type  = "AAAA"
	AliyunDNS = "dns.aliyuncs.com" // 修改DNS的API地址
)

// 配置文件
type Config struct {
	AccessKeyId     string `yaml:"accessKeyId"`
	AccessKeySecret string `yaml:"accessKeySecret"`
	TTL             int64  `yaml:"ttl"`
	IPv4            string `yaml:"ipv4_check_url"`
	IPv6            string `yaml:"ipv6_check_url"`
	Interval        int    `yaml:"interval"`
	BroadbandRetry  int8   `yaml:"broadband_retry"`
	UseCache        bool   `yaml:"cache"`
	Customer        []Customer `yaml:"customer"`
	IPsCache        map[string]string
	RecordIds       map[string]string // 记录DNS的recordID
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

	// 缓存处理
	if c.UseCache {
		c.IPsCache = make(map[string]string)
		c.RecordIds = make(map[string]string)
	} else {
		c.IPsCache = nil
		c.RecordIds = nil
	}

	return nil
}

// 初始化
func (c *Config) InitBroadbandRecords() map[string]bool {
	if c.BroadbandRetry > 0 {
		return make(map[string]bool, c.BroadbandRetry)
	}
	return nil
}

// 初始化IP缓存
func (c *Config) InitCache(cli *alidns.Client) {
	if !c.UseCache {
		fmt.Println("Cache is OFF and will not be used.")
		return
	}
	for _, customer := range c.Customer {
		if (customer.IPv4RR != "") || (c.IPv4 != "") {
			subDomain := customer.IPv4RR + "." + customer.Domain
			c.subDomainRecordsToCache(cli, &subDomain)
		}
		if (customer.IPv6RR != "") || (c.IPv6 != "") {
			subDomain := customer.IPv6RR + "." + customer.Domain
			c.subDomainRecordsToCache(cli, &subDomain)
		}
	}
}

// 查询DNS记录并更新到缓存
func (c *Config) subDomainRecordsToCache(cli *alidns.Client, subDomain *string) {
	subDomainRequest := &alidns.DescribeSubDomainRecordsRequest{
		SubDomain: subDomain,
	}
	subDomainRecords, err := cli.DescribeSubDomainRecords(subDomainRequest)
	if err == nil {
		// 获取域名解析记录，更新缓存
		for _, record := range subDomainRecords.Body.DomainRecords.Record {
			// www.yourdomain.com#A => 127.0.0.1
			// www.yourdomain.com#AAAA => ::1
			cacheKey := CacheKey(subDomain,record.Type)
			c.IPsCache[cacheKey] = *record.Value
			c.RecordIds[cacheKey] = *record.RecordId
		}
	}
}

func PublilcIPs(IPv4CheckUrl, IPv6CheckUrl string) (PubIPv4, PubIPv6 *string) {
	PubIPv4, PubIPv6 = nil, nil
	if IPv4CheckUrl != "" {
		PubIPv4 = GetPublishIP(IPv4CheckUrl)
	}
	if IPv6CheckUrl != "" {
		PubIPv6 = GetPublishIP(IPv6CheckUrl)
	}
	return PubIPv4, PubIPv6
}

func MultiBroadbandPublicIPs(IPv4CheckUrl, IPv6CheckUrl string, broadbandRetry int8) (map[string]bool, map[string]bool) {
	if broadbandRetry < 2 {
		return nil, nil
	}

	broadbandIPv4 := make(map[string]bool, 0)
	broadbandIPv6 := make(map[string]bool, 0)

	for i := int8(0); i < broadbandRetry; i++ {
		pubIPv4, pubIPv6 := PublilcIPs(IPv4CheckUrl, IPv6CheckUrl)
		if pubIPv4 != nil {
			broadbandIPv4[*pubIPv4] = true
		}
		if pubIPv6 != nil {
			broadbandIPv6[*pubIPv6] = true
		}
	}
	if len(broadbandIPv4) == 0 {
		broadbandIPv4 = nil
	}
	if len(broadbandIPv6) == 0 {
		broadbandIPv6 = nil
	}
	return broadbandIPv4, broadbandIPv6
}

func BroadbandIPFisrt(broadbandIPs map[string]bool) string {
	for broadbandIP := range broadbandIPs {
		return broadbandIP
	}
	return ""
}

func GetPublishIP(IPCheckUrl string) *string {
	resp, err := http.Get("http://" + IPCheckUrl)
	if err != nil {
		fmt.Printf("公网IP查询失败 ：%s\r\n", err.Error())
		return nil
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	ip1 := strings.Replace(string(content), "\n", "", -1)
	ip2 := strings.Replace(ip1, "\r", "", -1)
	return &ip2
}

func CacheKey(subDomain, ipType *string) string {
	return (*subDomain + "#" + *ipType)
}

// 判断ip是否发生变动
func DoesIPChanged(config *Config, broadbandIP map[string]bool, cacheKey, oldIP, newIP string) bool {
	if config.BroadbandRetry > 1 {
		if config.UseCache {
			// 缓存中的IP出现在本次公网查询结果，保持解析记录不变
			return broadbandIP[config.IPsCache[cacheKey]]
		} else {
			// 正在使用的IP命中查询结果集
			return broadbandIP[oldIP]
		}
	} else {
		if config.UseCache {
			// IP在缓存中，IP 没有发生变化，不做任何操作
			return config.IPsCache[cacheKey] == newIP
		} else {
			// IP 没有发生变化，不做任何操作
			return oldIP == newIP
		}
	}
}
