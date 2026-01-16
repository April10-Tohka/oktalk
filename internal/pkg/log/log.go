package log

import (
	"io"
	"oktalk/internal/pkg/config"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 定义常量
const TraceIDKey = "trace_id"

// InitLog 初始化日志配置
func InitLog(conf *config.Config) {
	// 1. 设置输出格式 (带颜色的文本)
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
		ForceColors:     true, // 强制开启颜色，让控制台更好看
		DisableColors:   false,
		PadLevelText:    true, // 对齐级别文本 (INFO, ERROR 等)
	}

	logrus.SetFormatter(formatter)

	// 2. 设置日志级别
	level, err := logrus.ParseLevel(conf.Server.Mode) // 根据模式决定级别
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// 3. 配置多端输出 (控制台 + 文件)
	fileWriter := &lumberjack.Logger{
		Filename:   "storage/logs/oktalk.log",
		MaxSize:    500, // MB
		MaxBackups: 10,
		MaxAge:     7, // Days
		Compress:   true,
	}

	// 同时输出到标准输出和文件
	multiWriter := io.MultiWriter(os.Stdout, fileWriter)
	logrus.SetOutput(multiWriter)
	logrus.AddHook(&TraceContextHook{})
	// 4. 开启行号报告
	logrus.SetReportCaller(true)
}
