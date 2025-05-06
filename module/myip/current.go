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
	alidns "github.com/alibabacloud-go/alidns-20150109/client"

	"aliyun-dns/config"
)

// 当前使用的ip的缓存
var CurrentCache = LocalCache{}

// ------------------------------------------------------------------------
//
// ------------------------------------------------------------------------

// 缓存当前正在使用的域名、ip关系
type LocalCache struct {
	//current  map[string]string // 记录当前正在使用域名和IP
	IP       map[string]string // 记录阿里云dns的ip
	RecordId map[string]string // 记录阿里云DNS的recordID
}

// 初始化IP缓存
func (c *LocalCache) Init(cfg *config.Config, cli alidns.Client) {
	//c.current = make(map[string]string)
	c.IP = make(map[string]string)
	c.RecordId = make(map[string]string)
	for _, customer := range cfg.Customer {
		if (customer.IPv4RR != "") || (cfg.IPv4 != "") {
			subDomain := customer.IPv4RR + "." + customer.Domain
			c.RecordsToCache(cli, &subDomain)
		}
		if (customer.IPv6RR != "") || (cfg.IPv6 != "") {
			subDomain := customer.IPv6RR + "." + customer.Domain
			c.RecordsToCache(cli, &subDomain)
		}
	}
}

// 添加或更新
func (c *LocalCache) Put(cacheKey, ip, recordId string) {
	//c.current[cacheKey] = ip
	c.IP[cacheKey] = ip
	c.RecordId[cacheKey] = recordId
}

// 获取缓存的ip
func (c *LocalCache) GetIp(cacheKey string) string {
	return c.IP[cacheKey]
}

// 获取缓存的recordid
func (c *LocalCache) GetRecordId(cacheKey string) string {
	return c.RecordId[cacheKey]
}

// 判断ip是否在缓存中
func (c *LocalCache) IsIPIn(cacheKey, ip string) bool {
	return c.IP[cacheKey] == ip
}

// 域名的缓存是否存在
func (c *LocalCache) IsNotExist(cacheKey string) bool {
	return c.IP[cacheKey] == ""
}

// 判断RecordId是否在缓存中
func (c *LocalCache) IsRecordIdIn(cacheKey, id string) bool {
	return c.RecordId[cacheKey] == id
}

// 删除记录
func (c *LocalCache) Del(cacheKey string) {
	delete(c.IP, cacheKey)
	delete(c.RecordId, cacheKey)
}

// 查询DNS记录并更新到缓存
func (c *LocalCache) RecordsToCache(cli alidns.Client, subDomain *string) {
	subDomainRequest := &alidns.DescribeSubDomainRecordsRequest{
		SubDomain: subDomain,
	}
	subDomainRecords, err := cli.DescribeSubDomainRecords(subDomainRequest)
	if err == nil {
		// 获取域名解析记录，更新缓存
		for _, record := range subDomainRecords.Body.DomainRecords.Record {
			// www.yourdomain.com#A => 127.0.0.1
			// www.yourdomain.com#AAAA => ::1
			cacheKey := CacheKey(*subDomain, *record.Type)
			c.IP[cacheKey] = *record.Value
			c.RecordId[cacheKey] = *record.RecordId
		}
	}
}

// 判断ip是否发生变动
// 通过nslooup查询当前域名的IP，并和新的公网IP做对比
func DoesIPChanged(broadbandIP map[string]bool, cacheKey, ip string) bool {
	if len(broadbandIP) > 0 && broadbandIP[CurrentCache.GetIp(cacheKey)] {
		// 宽带多播时，通过key在缓存中查询ip，并与当前多播结果对比
		return false
	} else {
		// 缓存中的IP出现在本次公网查询结果，保持解析记录不变
		return !CurrentCache.IsIPIn(cacheKey, ip)
	}
}
