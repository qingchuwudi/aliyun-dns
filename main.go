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
	myConfig "aliyun-dns/config"
	"flag"
	"fmt"
	alidns "github.com/alibabacloud-go/alidns-20150109/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
	"os"
	"time"
)

var (
	h bool
	c string
)

func init() {
	flag.BoolVar(&h, "h", false, "This help")
	flag.StringVar(&c, "c", "", "Full path of the config file.")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `DDNS with Aliyun dns written in go.
Usage: aliyun-ddns [-h] [-c '/full/path/of/config.yaml']

Options:
`)
	flag.PrintDefaults()
}



// 创建客户端
func CreateClient(accessKeyId, accessKeySecret *string) (_result *alidns.Client, _err error) {
	config := &openapi.Config{}
	config.AccessKeyId = accessKeyId
	config.AccessKeySecret = accessKeySecret
	config.Endpoint = tea.String(myConfig.AliyunDNS)
	_result = &alidns.Client{}
	_result, _err = alidns.NewClient(config)
	return _result, _err
}

func run(config *myConfig.Config, cli *alidns.Client) {
	// 查询IP
	PubIPv4, PubIPv6 := myConfig.PublilcIPs(config.IPv4, config.IPv6)

	// 更新所有用户配置
	for _, customer := range config.Customer {
		if (PubIPv4 != nil) && (customer.IPv4RR != "") {
			err := updateDomains(cli, config, nil, customer.IPv4RR, customer.Domain, *PubIPv4, myConfig.IPv4Type, config.TTL)
			if err != nil {
				fmt.Printf("update IPv4 failed : %s\r\n", err.Error())
			}
		}
		if (PubIPv6 != nil) && (customer.IPv6RR != "") {
			err := updateDomains(cli, config, nil, customer.IPv6RR, customer.Domain, *PubIPv6, myConfig.IPv6Type, config.TTL)
			if err != nil {
				fmt.Printf("update IPv6 failed : %s\r\n", err.Error())
			}
		}
	}
}

func runMultiBroadband(config *myConfig.Config, cli *alidns.Client) {
	if config.BroadbandRetry < 2 {
		fmt.Println("broadband_retry must be bigger than 1.")
		return
	}
	// 重复获取公网IP
	broadbandIPv4, broadbandIPv6 := myConfig.MultiBroadbandPublicIPs(config.IPv4, config.IPv6, config.BroadbandRetry)

	// 更新所有用户配置
	for _, customer := range config.Customer {
		if broadbandIPv4 != nil {
			IP := myConfig.BroadbandIPFisrt(broadbandIPv4)
			err := updateDomains(cli, config, broadbandIPv4, customer.IPv4RR, customer.Domain, IP, myConfig.IPv4Type, config.TTL)
			if err != nil {
				fmt.Printf("update IPv4 failed : %s\r\n", err.Error())
			}
		}
		if broadbandIPv6 != nil {
			IP := myConfig.BroadbandIPFisrt(broadbandIPv6)
			err := updateDomains(cli, config, broadbandIPv6, customer.IPv6RR, customer.Domain, IP, myConfig.IPv6Type, config.TTL)
			if err != nil {
				fmt.Printf("update IPv6 failed : %s\r\n", err.Error())
			}
		}
	}
}


func updateDomains(cli *alidns.Client, config *myConfig.Config, broadbandIP map[string]bool, DomainRR, Domain, IP, ipType string, ttl int64) error {
	subDomain := DomainRR + "." + Domain
	// 使用缓存
	if config.UseCache {
		cacheKey := myConfig.CacheKey(&subDomain, &ipType)
		if myConfig.DoesIPChanged(config, broadbandIP, cacheKey, "", IP){
			return nil
		} else if config.IPsCache[cacheKey] == "" {
			// 没有记录
			AddDomainRecordRequest := &alidns.AddDomainRecordRequest{
				DomainName: &Domain,
				RR:         &DomainRR,
				Type:       &ipType,
				Value:      &IP,
				TTL:        &ttl,
			}
			resp, err := cli.AddDomainRecord(AddDomainRecordRequest)
			if err != nil {
				return err
			}
			fmt.Printf("[%s] - New public IP : %s => %s\r\n", time.Now().Format("2006-01-02 15:04:05"), cacheKey, IP)
			// 更新缓存
			config.IPsCache[cacheKey] = IP
			config.RecordIds[cacheKey] = *resp.Body.RecordId
			return nil
		}
		// 有记录，更新
		recordId := config.RecordIds[cacheKey]
		updateDomainRecordRequest := &alidns.UpdateDomainRecordRequest{
			RecordId: &recordId,
			RR:       &DomainRR,
			Type:     &ipType,
			Value:    &IP,
		}
		_, err := cli.UpdateDomainRecord(updateDomainRecordRequest)
		if err != nil {
			return err
		}
		config.IPsCache[cacheKey] = IP // 更新缓存
		fmt.Printf("[%s] - New public IP : %s => %s\r\n", time.Now().Format("2006-01-02 15:04:05"), cacheKey, IP)
		return nil
	} else {
		// 查询域名解析记录
		subDomainRequest := &alidns.DescribeSubDomainRecordsRequest{
			SubDomain: &subDomain,
		}
		subDomainRecords, err := cli.DescribeSubDomainRecords(subDomainRequest)
		if err != nil {
			return err
		}

		// 获取域名id，找到后更新，找不到添加
		for _, record := range subDomainRecords.Body.DomainRecords.Record {
			// 类型一致才可以修改
			if (*record.Type) == ipType {
				if myConfig.DoesIPChanged(config, broadbandIP, "", *record.Value, IP) {
					return nil
				}
				fmt.Printf("[%s] - New public IP : %s\r\n", time.Now().Format("2006-01-02 15:04:05"), IP)
				// 修改域名解析记录
				updateDomainRecordRequest := &alidns.UpdateDomainRecordRequest{
					RecordId: record.RecordId,
					RR:       record.RR,
					Type:     record.Type,
					Value:    &IP,
				}
				_, err = cli.UpdateDomainRecord(updateDomainRecordRequest)
				if err != nil {
					return err
				}
				return nil
			}
		}
		// 没有找到Type一致的记录
		AddDomainRecordRequest := &alidns.AddDomainRecordRequest{
			DomainName: &Domain,
			RR:         &DomainRR,
			Type:       &ipType,
			Value:      &IP,
			TTL:        &ttl,
		}
		_, err = cli.AddDomainRecord(AddDomainRecordRequest)
		if err != nil {
			return err
		}
		fmt.Printf("[%s] - New public IP : %s\r\n", time.Now().Format("2006-01-02 15:04:05"), IP)
		return nil
	}
}

func main() {
	flag.Parse()
	if h {
		flag.Usage()
		return
	}
	if c == "" {
		flag.Usage()
		return
	}
	// 读取配置
	var config myConfig.Config
	err := config.InitConfig(c)
	if err != nil {
		fmt.Printf("Parse config file error : %s\r\n", err.Error())
		return
	}

	// 创建客户端
	cli, err := CreateClient(tea.String(config.AccessKeyId), tea.String(config.AccessKeySecret))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 加载缓存
	config.InitCache(cli)

	interval := time.Second * time.Duration(config.Interval)
	// 根据宽带多拨情况来判断使用哪个函数主体运行
	var runFunc func(config2 *myConfig.Config, client *alidns.Client)
	if config.BroadbandRetry < 2 {
		fmt.Println("正常DDNS")
		runFunc = run
	} else {
		fmt.Println("宽带多拨DDNS")
		runFunc = runMultiBroadband
	}

	for {
		runFunc(&config, cli)
		time.Sleep(interval)
	}
}
