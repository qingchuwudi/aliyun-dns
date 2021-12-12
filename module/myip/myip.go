/*
 * Copyright (c) 2021 qingchuwudi
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * author bypf2009@vip.qq.com
 * create at 2021/12/10
 */

package myip

import (
	"io/ioutil"
	"net/http"
	"strings"

	"aliyun-dns/module/loger"
)

// 获取公网IP
func PublilcIPs(IPv4CheckUrl, IPv6CheckUrl string) (PubIPv4, PubIPv6 string) {
	PubIPv4, PubIPv6 = "", ""
	if IPv4CheckUrl != "" {
		PubIPv4 = GetPublishIP(IPv4CheckUrl)
	}
	if IPv6CheckUrl != "" {
		PubIPv6 = GetPublishIP(IPv6CheckUrl)
	}
	return PubIPv4, PubIPv6
}

// 多拨情况下获取多个公网IP
func MultiBroadbandPublicIPs(IPv4CheckUrl, IPv6CheckUrl string, broadbandRetry int8) (map[string]bool, map[string]bool) {
	if broadbandRetry < 2 {
		loger.Info("要运行多拨环境，配置文件中的参数 'broadband_retry' 必须大于 1")
		return nil, nil
	}

	broadbandIPv4 := make(map[string]bool, 0)
	broadbandIPv6 := make(map[string]bool, 0)

	for i := int8(0); i < broadbandRetry; i++ {
		pubIPv4, pubIPv6 := PublilcIPs(IPv4CheckUrl, IPv6CheckUrl)
		if pubIPv4 != "" {
			broadbandIPv4[pubIPv4] = true
		}
		if pubIPv6 != "" {
			broadbandIPv6[pubIPv6] = true
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

// 获取IP地址
func GetPublishIP(IPCheckUrl string) string {
	resp, err := http.Get("http://" + IPCheckUrl)
	if err != nil {
		loger.Info("公网IP查询失败 ：%s", err.Error())
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	ip1 := strings.Replace(string(content), "\n", "", -1)
	ip2 := strings.Replace(ip1, "\r", "", -1)
	return ip2
}

// 构造缓存的key
func CacheKey(subDomain, ipType string) string {
	return (subDomain + "#" + ipType)
}
