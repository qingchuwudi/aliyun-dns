/*
 * Copyright (C) 2021 qingchuwudi
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

package help

import (
	"flag"
	"fmt"
	"os"

	"aliyun-dns/module/filecheck"
	"aliyun-dns/module/loger"
)

var (
	Help bool
	Ver  bool
	Cfg  string
)

func init() {
	flag.BoolVar(&Help, "h", false, "查看帮助")
	flag.BoolVar(&Ver, "v", false, "查看软件版本")
	flag.StringVar(&Cfg, "c", "", "配置文件的完整路径（绝对路径）。例如: -c /etc/aliyun-dns.yaml")
	flag.Usage = Usage
}

func Usage() {
	fmt.Fprintf(os.Stderr, `基于阿里云平台的DDNS（域名解析自动更新）工具.
Usage: aliyun-ddns [-h | -v | -c ]

Options:
`)
	flag.PrintDefaults()
}

// 获取命令行参数（读取并解析）
func ParseArgs() (stop bool) {
	flag.Parse()
	if Help {
		flag.Usage()
		return true
	}
	if Ver {
		fmt.Println("v2.1.0")
		return true
	}
	if Cfg == "" {
		loger.PreError("请指定配置文件")
		return true
	}

	// 检查文件是否存在
	if !filecheck.IsFileValid(Cfg) {
		loger.PreError("配置文件不存在，或者没有权限")
		return true
	}
	return false
}
