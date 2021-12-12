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
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ZapLoger *zap.Logger
var Loger *zap.SugaredLogger
var (
	// Debug = Loger.Debugf // debug日志
	// Info  = Loger.Infof  //
	// Warn  = Loger.Warnf
	// Error = Loger.Errorf
	// Panic = Loger.Panicf
	Debug func(template string, args ...interface{}) // debug日志
	Info  func(template string, args ...interface{}) //
	Warn  func(template string, args ...interface{})
	Error func(template string, args ...interface{})
	Fatal func(template string, args ...interface{})
	Panic func(template string, args ...interface{})
)

func InitLogger(logPath, filename string) {
	// 日志配置
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:    "msg",
		LevelKey:      "level",
		TimeKey:       "time",
		NameKey:       "logger",
		CallerKey:     "file",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		// EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	// 增加组件配置
	logfile := filepath.Join(logPath, filename)

	hook := lumberjack.Logger{
		Filename:   logfile, // 日志文件路径
		MaxSize:    16,     // 每个日志文件保存的大小 单位:M
		MaxAge:     7,       // 文件最多保存多少天
		MaxBackups: 10,      // 日志文件最多保存多少个备份
		Compress:   true,   // 是否压缩
	}

	var writes = []zapcore.WriteSyncer{zapcore.AddSync(&hook)}
	// 如果是开发环境，同时在控制台上也输出
	writes = append(writes, zapcore.AddSync(os.Stdout))
	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.DebugLevel)

	// 配置生效
	core := zapcore.NewCore(
		// zapcore.NewJSONEncoder(encoderConfig),
		// 日志格式默认是Json格式，转为普通格式的日志
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(writes...),
		atomicLevel,
	)

	development := zap.Development()

	// 构造日志
	ZapLoger = zap.New(core, development)
	ZapLoger.Info("服务启动，日志记录器启动成功")

	// 赋值
	Loger = ZapLoger.Sugar()
	Debug = Loger.Debugf // debug日志
	Info = Loger.Infof   //
	Warn = Loger.Warnf
	Error = Loger.Errorf
	Fatal = Loger.Fatalf
	Panic = Loger.Panicf
}
