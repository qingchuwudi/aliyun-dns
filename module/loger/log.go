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

// 日志配置
type LogConfig struct {
    Path    string `json:"path,omitempty"`    // 日志文件路径
    Level   string `json:"level,omitempty"`   // 记录的日志等级：debug,info,warn,error
    Develop bool   `json:"develop,omitempty"` // 开发者模式，开启后会输出代码文件和堆栈信息
}

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

func InitLogger(cfg *LogConfig) {
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
    hook := getWriter(cfg.Path)
    var writes = []zapcore.WriteSyncer{zapcore.AddSync(hook)}
    // 同时在控制台上也输出
    writes = append(writes, zapcore.AddSync(os.Stdout))
    // 设置日志级别
    atomicLevel := zap.NewAtomicLevel()
    // atomicLevel.SetLevel(zap.DebugLevel)
    atomicLevel.SetLevel(getLevel(cfg.Level))

    // 配置生效
    core := zapcore.NewCore(
        // zapcore.NewJSONEncoder(encoderConfig),
        // 日志格式默认是Json格式，转为普通格式的日志
        zapcore.NewConsoleEncoder(encoderConfig),
        zapcore.NewMultiWriteSyncer(writes...),
        atomicLevel,
    )
    if cfg.Develop {
        // 开启开发模式，堆栈跟踪(可以看到文件名、代码行数)
        // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数
        caller := zap.AddCaller()
        // 开启文件及行号
        development := zap.Development()
        // 构造日志
        ZapLoger = zap.New(core, caller, development)
    } else {
        ZapLoger = zap.New(core)
    }

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

// 日志等级
func getLevel(level string) zapcore.Level {
    switch level {
    case "debug":
        return zap.DebugLevel
    case "info":
        return zap.InfoLevel
    case "warn":
        return zap.WarnLevel
    case "error":
        return zap.ErrorLevel
    case "panic":
        return zap.PanicLevel
    case "fatal":
        return zap.FatalLevel
    default:
        return zap.InfoLevel
    }
}

// 日志筛选器
func levelFilter(level zapcore.Level) func(zapcore.Level) bool {
    max := zapcore.FatalLevel + 1
    min := level - 1
    return func(lvl zapcore.Level) bool {
        return lvl < max && lvl > min
    }
}

// 获取日志输出文件
func getWriter(logPath string) *lumberjack.Logger {
    fullfile := filepath.Join(logPath, "aliyun-dns.log")
    return &lumberjack.Logger{
        Filename:   fullfile, // 日志文件路径
        MaxSize:    16,       // 每个日志文件保存的大小 单位:M
        MaxAge:     30,       // 文件最多保存多少天
        MaxBackups: 30,       // 日志文件最多保存多少个备份
        LocalTime:  true,     // 使用本地时间记录
        Compress:   true,     // 是否压缩
    }
}
