/*
 * Copyright 2021-2021 qingchuwudi
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

package filecheck

import (
    "os"

    "aliyun-dns/module/loger"
)

// 判断所给路径文件/文件夹是否存在
func IsFileValid(file string) bool {
    stu, err := os.Stat(file) // os.Stat获取文件信息
    if err != nil {
        if os.IsExist(err) {
            return true
        }
        loger.PreError("文件 '%s' 不存在或没有权限。", file)
        return false
    }
    if stu.IsDir() {
        loger.PreError("参数错误，'%s' 是文件夹。", file)
        return false
    }
    return true
}

// 判断所给路径文件/文件夹是否存在
func IsExist(file string) bool {
    _, err := os.Stat(file) // os.Stat获取文件信息
    return os.IsExist(err)
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
    s, err := os.Stat(path)
    if err != nil {
        return false
    }
    return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
    return !IsDir(path)
}
