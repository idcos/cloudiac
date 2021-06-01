package logs

import (
	"fmt"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
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

// 根据配置文件配置 logger
func Init(level string, maxDays int, name string) {
	if maxDays == 0 {
		maxDays = 7
	}
	abs, _ := filepath.Abs(os.Args[0])
	dir := filepath.Dir(abs)
	ext := filepath.Ext(name)
	execName := name[:len(name)-len(ext)]

	logPath := filepath.Join(dir, "logs", execName+".log")
	//日志按天切割
	rl, err := rotatelogs.New(
		fmt.Sprintf("%s.%%Y%%m%%d", logPath),
		rotatelogs.WithLinkName(logPath),
		rotatelogs.WithRotationTime(time.Hour*24),
		rotatelogs.WithMaxAge(time.Hour*24*time.Duration(maxDays)),
	)
	if err != nil {
		logrus.Fatalf("failed to open log file: %v", err)
	}
	// 同时进行标准输出
	writers := io.MultiWriter(rl, os.Stdout)

	defaultLogger.SetOutput(writers)
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

//func init() {
//	Init("DEBUG")
//}
