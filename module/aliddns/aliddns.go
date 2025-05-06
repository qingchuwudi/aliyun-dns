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
	"context"

	"github.com/alibabacloud-go/alidns-20150109/client"
	aliCliSdk "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
	"go.uber.org/zap"

	myConfig "aliyun-dns/config"
	"aliyun-dns/module/loger"
	"aliyun-dns/module/myip"
)

// 创建客户端
func CreateClient(accessKeyId, accessKeySecret *string) (*client.Client, error) {
	cfg := &aliCliSdk.Config{}
	cfg.AccessKeyId = accessKeyId
	cfg.AccessKeySecret = accessKeySecret
	cfg.Endpoint = tea.String(myConfig.AliyunDNS)
	return client.NewClient(cfg)
}

// 正常宽带
func Run(ctx context.Context, cfg *myConfig.Config, cli client.Client) {
	// 查询IP
	PubIPv4, PubIPv6 := myip.PublilcIPs(cfg.IPv4, cfg.IPv6)

	// 更新所有用户配置
	for _, customer := range cfg.Customer {
		select {
		case <-ctx.Done():
			return
		default:
			if (PubIPv4 != "") && (customer.IPv4RR != "") {
				err := UpdateDomains(cli, nil, customer.IPv4RR, customer.Domain, PubIPv4, myConfig.IPv4Type, cfg.TTL)
				if err != nil {
					loger.Error("IPv4 update failed", zap.Error(err))
				}
			}
			if (PubIPv6 != "") && (customer.IPv6RR != "") {
				err := UpdateDomains(cli, nil, customer.IPv6RR, customer.Domain, PubIPv6, myConfig.IPv6Type, cfg.TTL)
				if err != nil {
					loger.Error("IPv6 update failed", zap.Error(err))
				}
			}
		}
	}
}

// 宽带多拨或有多条宽带线路
func RunOnMultiBroadband(ctx context.Context, cfg *myConfig.Config, cli client.Client) {
	// 重复获取公网IP
	broadbandIPv4, broadbandIPv6 := myip.MultiBroadbandPublicIPs(cfg.IPv4, cfg.IPv6, cfg.BroadbandRetry)

	// 更新所有用户配置
	for _, customer := range cfg.Customer {
		select {
		case <-ctx.Done():
			return
		default:
			if broadbandIPv4 != nil {
				IP := myip.BroadbandIPFisrt(broadbandIPv4)
				err := UpdateDomains(cli, broadbandIPv4, customer.IPv4RR, customer.Domain, IP, myConfig.IPv4Type, cfg.TTL)
				if err != nil {
					loger.Error("update IPv4 failed", zap.Error(err))
				}
			}
			if broadbandIPv6 != nil {
				IP := myip.BroadbandIPFisrt(broadbandIPv6)
				err := UpdateDomains(cli, broadbandIPv6, customer.IPv6RR, customer.Domain, IP, myConfig.IPv6Type, cfg.TTL)
				if err != nil {
					loger.Error("update IPv6 failed", zap.Error(err))
				}
			}
		}
	}
}
