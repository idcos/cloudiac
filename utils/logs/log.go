package logs

import (
	"fmt"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
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
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000",
	})
	return logger
}

func init() {
	defaultLogger = newLogger()
}

// Init 根据配置文件配置 logger
func Init(level string, path string, maxDays int) {
	writers := []io.Writer{os.Stdout}

	if path != "" {
		if maxDays == 0 {
			maxDays = 7
		}
		//日志按天切割
		rl, err := rotatelogs.New(
			fmt.Sprintf("%s.%%Y%%m%%d", path),
			rotatelogs.WithLinkName(path),
			rotatelogs.WithRotationTime(time.Hour*24),
			rotatelogs.WithMaxAge(time.Hour*24*time.Duration(maxDays)),
		)
		if err != nil {
			logrus.Fatalf("failed to open log file: %v", err)
		} else {
			writers = append(writers, rl)
		}
	}

	defaultLogger.SetOutput(io.MultiWriter(writers...))
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
