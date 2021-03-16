package logs

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	logrus.Ext1FieldLogger
}

var (
	defaultLogger *logrus.Logger
)

// 创建一个新的默认 logger
func newLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000",
	})
	return logger
}

func init() {
	defaultLogger = newLogger()
}

// 根据配置文件配置 logger
func Init(level string) {
	if level != "" {
		lvl, err := logrus.ParseLevel(level)
		if err != nil {
			defaultLogger.Fatalln(err)
		}
		defaultLogger.SetLevel(lvl)
	}
}

func Get() Logger {
	return defaultLogger
}

func init() {
	Init("DEBUG")
}
