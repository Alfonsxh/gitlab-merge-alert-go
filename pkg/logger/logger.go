package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func Init(level string) {
	log = logrus.New()

	// 确保日志目录存在
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.WithError(err).Warn("Failed to create log directory, using stdout only")
		log.SetOutput(os.Stdout)
	} else {
		// 创建日志文件
		logFile := filepath.Join(logDir, "app.log")
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.WithError(err).Warn("Failed to open log file, using stdout only")
			log.SetOutput(os.Stdout)
		} else {
			// 同时输出到文件和标准输出
			multiWriter := io.MultiWriter(os.Stdout, file)
			log.SetOutput(multiWriter)
		}
	}

	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	// 设置日志级别
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

func GetLogger() *logrus.Logger {
	if log == nil {
		Init("info")
	}
	return log
}
