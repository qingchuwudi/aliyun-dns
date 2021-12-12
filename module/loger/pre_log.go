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

package loger

import (
    "fmt"
)

// ------------------------------------------------------------------
// 透明背景的输出
// ------------------------------------------------------------------

func PreSuccess(format string, a ...interface{}) {
    fmt.Printf("\033[1;32;48m"+format+"\033[0m\n", a...)
}

func PreInfo(format string, a ...interface{}) {
    fmt.Printf("\033[1;37;48m"+format+"\033[0m\n", a...)
}

func PreError(format string, a ...interface{}) {
    fmt.Printf("\033[1;31;48m"+format+"\033[0m\n", a...)
}

// ------------------------------------------------------------------
// 带背景的输出
// ------------------------------------------------------------------

// 背景
func PreSuccessHeav(format string, a ...interface{}) {
    fmt.Printf("\033[1;32;47m"+format+"\033[0m\n", a...)
}

// 背景黄色
func PreInfoHeav(format string, a ...interface{}) {
    fmt.Printf("\033[1;37;43m"+format+"\033[0m\n", a...)
}

// 重量级错误提示，背景黑色
func PreErrorHeav(format string, a ...interface{}) {
    fmt.Printf("\033[1;31;40m"+format+"\033[0m\n", a...)
}
