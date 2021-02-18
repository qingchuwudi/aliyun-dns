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
    "crypto/tls"
    "flag"
    "fmt"
    alidns "github.com/alibabacloud-go/alidns-20150109/client"
    openapi "github.com/alibabacloud-go/darabonba-openapi/client"
    "github.com/alibabacloud-go/tea/tea"
    "net/http"
    "os"
    "time"
)


var tr = &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

// 实际中应该用更好的变量名
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
func CreateClient(accessKeyId *string, accessKeySecret *string) (_result *alidns.Client, _err error) {
    config := &openapi.Config{}
    config.AccessKeyId = accessKeyId
    config.AccessKeySecret = accessKeySecret
    config.Endpoint = tea.String(myConfig.AliyunDNS)
    _result = &alidns.Client{}
    _result, _err = alidns.NewClient(config)
    return _result, _err
}

func run(access myConfig.Config) {
    cli, err := CreateClient(tea.String(access.AccessKeyId), tea.String(access.AccessKeySecret))
    if err != nil {
        fmt.Println(err.Error())
        return
    }

    // 查询IP
    PubIPv4, PubIPv6 := access.PublilcIPs()

    // 更新所有用户配置
    for _, customer := range access.Customer {
        if PubIPv4 != nil {
            err = updateDomains(cli , customer.IPv4RR, customer.Domain, *PubIPv4, myConfig.IPv4Type, access.TTL)
            if err != nil {
                fmt.Printf("update IPv4 failed : %s\r\n", err.Error())
            }
        }
        if PubIPv6 != nil {
            err = updateDomains(cli , customer.IPv6RR, customer.Domain, *PubIPv6, myConfig.IPv6Type, access.TTL)
            if err != nil {
                fmt.Printf("update IPv6 failed : %s\r\n", err.Error())
            }
        }
    }
}

func updateDomains(cli *alidns.Client,DomainRR, Domain, IP string, ipType string, ttl int64) error {
    // 查询域名解析记录
    subDomain := DomainRR + "." + Domain
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
            // IP 没有发生变化，不做任何操作
            if (*record.Value) == IP {
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
        DomainName:   &Domain,
        RR:           &DomainRR,
        Type:         &ipType,
        Value:        &IP,
        TTL:          &ttl,
    }
    _, err = cli.AddDomainRecord(AddDomainRecordRequest)
    if err != nil {
        return err
    }
    return nil
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

    for {
        run(config)
        time.Sleep(time.Second * time.Duration(config.Interval))
    }
}
