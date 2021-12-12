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

package aliddns

import (
	"strings"

	"aliyun-dns/module/loger"
	"aliyun-dns/module/myip"
	"github.com/alibabacloud-go/alidns-20150109/client"
)

// 更新域名配置
//func UpdateDomains(cli *client.Client, cfg *config.Config, broadbandIP map[string]bool, DomainRR, Domain, IP, ipType string, ttl int64) error {
//	subDomain := DomainRR + "." + Domain
//	// 使用缓存
//	if cfg.UseCache {
//		cacheKey := myip.CacheKey(subDomain, ipType)
//		if myip.DoesIPChanged(broadbandIP, cacheKey, IP) {
//			return nil
//		} else if cfg.IPsCache[cacheKey] == "" {
//			// 没有记录
//			AddDomainRecordRequest := &client.AddDomainRecordRequest{
//				DomainName: &Domain,
//				RR:         &DomainRR,
//				Type:       &ipType,
//				Value:      &IP,
//				TTL:        &ttl,
//			}
//			resp, err := cli.AddDomainRecord(AddDomainRecordRequest)
//			if err != nil {
//				if ErrorDomainRecordDuplicate(err) {
//					// 更新缓存
//					cfg.IPsCache[cacheKey] = IP
//					cfg.RecordIds[cacheKey] = *resp.Body.RecordId
//					loger.Info("添加：已经有记录，更新缓存并跳过添加过程")
//					return nil
//				}
//				return err
//			}
//			// 更新缓存
//			cfg.IPsCache[cacheKey] = IP
//			cfg.RecordIds[cacheKey] = *resp.Body.RecordId
//			loger.Info("[%s] 公网IP已添加: %s", subDomain, IP)
//			return nil
//		}
//		// 有记录，更新
//		recordId := cfg.RecordIds[cacheKey]
//		updateDomainRecordRequest := &client.UpdateDomainRecordRequest{
//			RecordId: &recordId,
//			RR:       &DomainRR,
//			Type:     &ipType,
//			Value:    &IP,
//		}
//		resp, err := cli.UpdateDomainRecord(updateDomainRecordRequest)
//		if err != nil {
//			if ErrorDomainRecordDuplicate(err) {
//				// 更新缓存
//				cfg.IPsCache[cacheKey] = IP
//				cfg.RecordIds[cacheKey] = *resp.Body.RecordId
//				loger.Info("更新：已经有相同记录，稍后重试过程")
//				return nil
//			}
//			return err
//		}
//
//		cfg.IPsCache[cacheKey] = IP // 更新缓存
//
//		loger.Info("[%s] 公网IP更新: %s", subDomain, IP)
//		return nil
//	} else {
//		// 查询域名解析记录
//		subDomainRequest := &client.DescribeSubDomainRecordsRequest{
//			SubDomain: &subDomain,
//		}
//		subDomainRecords, err := cli.DescribeSubDomainRecords(subDomainRequest)
//		if err != nil {
//			return err
//		}
//
//		// 获取域名id，找到后更新，找不到添加
//		for _, record := range subDomainRecords.Body.DomainRecords.Record {
//			// 类型一致才可以修改
//			if (*record.Type) == ipType {
//				myip.CurrentCache.Put(myip.CacheKey(subDomain, ipType), *record.Value)
//				if myip.DoesIPChanged(broadbandIP, subDomain, IP) {
//					return nil
//				}
//
//				// 修改域名解析记录
//				updateDomainRecordRequest := &client.UpdateDomainRecordRequest{
//					RecordId: record.RecordId,
//					RR:       record.RR,
//					Type:     record.Type,
//					Value:    &IP,
//				}
//				_, err = cli.UpdateDomainRecord(updateDomainRecordRequest)
//				if err != nil {
//					if ErrorDomainRecordDuplicate(err) {
//						loger.Info("更新：已经有相同记录，稍后重试过程")
//						return nil
//					}
//					return err
//				}
//				loger.Info("[%s] 公网IP更新: %s", subDomain, IP)
//				return nil
//			}
//		}
//		// 没有找到Type一致的记录
//		AddDomainRecordRequest := &client.AddDomainRecordRequest{
//			DomainName: &Domain,
//			RR:         &DomainRR,
//			Type:       &ipType,
//			Value:      &IP,
//			TTL:        &ttl,
//		}
//		_, err = cli.AddDomainRecord(AddDomainRecordRequest)
//		if err != nil {
//			if ErrorDomainRecordDuplicate(err) {
//				loger.Info("添加：已经有记录，跳过添加过程")
//				return nil
//			}
//			return err
//		}
//		loger.Info("[%s] 公网IP更新: %s", subDomain, IP)
//		return nil
//	}
//}

// 更新ip
func UpdateDomains(cli *client.Client, broadbandIP map[string]bool, DomainRR, Domain, IP, ipType string, ttl int64) error {
	subDomain := DomainRR + "." + Domain
	cacheKey := myip.CacheKey(subDomain, ipType)
	if !myip.DoesIPChanged(broadbandIP, cacheKey, IP) {
		// 没变化，返回
		loger.Debug("[%s] IP没有发生变化", cacheKey)
		return nil
	}

	// 有变化： 并且缓存中没有有记录
	if myip.CurrentCache.IsNotExist(cacheKey) {
		AddDomainRecordRequest := &client.AddDomainRecordRequest{
			DomainName: &Domain,
			RR:         &DomainRR,
			Type:       &ipType,
			Value:      &IP,
			TTL:        &ttl,
		}
		resp, err := cli.AddDomainRecord(AddDomainRecordRequest)
		if err != nil {
			if ErrorDomainRecordDuplicate(err) {
				// 更新缓存
				myip.CurrentCache.Put(cacheKey, IP, *resp.Body.RecordId)
				loger.Info("添加：已经有记录，更新缓存并跳过添加过程")
				return nil
			}
			return err
		}
		// 更新缓存
		myip.CurrentCache.Put(cacheKey, IP, *resp.Body.RecordId)
		loger.Info("添加：[%s] 公网IP已添加: %s", cacheKey, IP)
	} else {
		// 有记录，更新
		recordId := myip.CurrentCache.GetRecordId(cacheKey)
		updateDomainRecordRequest := &client.UpdateDomainRecordRequest{
			RecordId: &recordId,
			RR:       &DomainRR,
			Type:     &ipType,
			Value:    &IP,
		}
		resp, err := cli.UpdateDomainRecord(updateDomainRecordRequest)
		if err != nil {
			if ErrorDomainRecordDuplicate(err) {
				// 更新缓存
				myip.CurrentCache.Put(cacheKey, IP, *resp.Body.RecordId)
				loger.Info("更新：已经有相同记录，稍后重试")
				return nil
			}
			return err
		}

		myip.CurrentCache.Put(cacheKey, IP, *resp.Body.RecordId)

		loger.Info("[%s] 公网IP更新: %s", cacheKey, IP)
	}

	return nil
}

// 错误处理
func ErrorDomainRecordDuplicate(err error) bool {
	return strings.Contains(err.Error(), "DomainRecordDuplicate")
}
