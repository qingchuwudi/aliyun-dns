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

	"github.com/alibabacloud-go/alidns-20150109/client"
	"go.uber.org/zap"

	"aliyun-dns/module/loger"
	"aliyun-dns/module/myip"
)

// 更新ip
func UpdateDomains(cli client.Client, broadbandIP map[string]bool, DomainRR, Domain, IP, ipType string, ttl int64) error {
	subDomain := DomainRR + "." + Domain
	cacheKey := myip.CacheKey(subDomain, ipType)
	if !myip.DoesIPChanged(broadbandIP, cacheKey, IP) {
		// 没变化，返回
		loger.Debug("ip没有发生变化", zap.String(cacheKey, IP))
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
				if resp != nil && resp.Body != nil {
					// 更新缓存
					myip.CurrentCache.Put(cacheKey, IP, *resp.Body.RecordId)
					loger.Info("添加：已经有记录，更新缓存并跳过添加过程")
				}
				return nil
			}
			return err
		}
		if resp != nil && resp.Body != nil {
			// 更新缓存
			myip.CurrentCache.Put(cacheKey, IP, *resp.Body.RecordId)
			loger.Info("公网IP已添加", zap.String(cacheKey, IP))
		}
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
				if resp != nil && resp.Body != nil {
					// 更新缓存
					myip.CurrentCache.Put(cacheKey, IP, *resp.Body.RecordId)
					loger.Info("已经有相同记录，稍后重试")
				}
				return nil
			}
			return err
		}
		if resp != nil && resp.Body != nil {
			// 更新缓存
			myip.CurrentCache.Put(cacheKey, IP, *resp.Body.RecordId)
			loger.Info("公网IP更新", zap.String(cacheKey, IP))
		}
	}

	return nil
}

// 错误处理
func ErrorDomainRecordDuplicate(err error) bool {
	return strings.Contains(err.Error(), "DomainRecordDuplicate")
}
