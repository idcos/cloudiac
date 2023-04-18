// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package logs

import (
	"fmt"
	"io"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

type Logger interface {
	logrus.Ext1FieldLogger
}

var (
	defaultLogger *logrus.Logger
	logWriter     io.Writer
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

func initWriter(logPath string, maxDays int) {
	writers := []io.Writer{os.Stdout}

	if logPath != "" {
		if maxDays == 0 {
			maxDays = 7
		}
		//日志按天切割
		rl, err := rotatelogs.New(
			fmt.Sprintf("%s.%%Y%%m%%d", logPath),
			rotatelogs.WithLinkName(logPath),
			rotatelogs.WithRotationTime(time.Hour*24),
			rotatelogs.WithMaxAge(time.Hour*24*time.Duration(maxDays)),
		)
		if err != nil {
			logrus.Fatalf("failed to open log file: %v", err)
		} else {
			writers = append(writers, rl)
		}
	}
	logWriter = io.MultiWriter(writers...)
}

// Init 根据配置文件配置 logger
func Init(level string, path string, maxDays int) {
	initWriter(path, maxDays)
	defaultLogger.SetOutput(logWriter)
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

func Writer() io.Writer {
	return logWriter
}

type LogWriter struct {
	write func(string)
}

func (w *LogWriter) Write(bytes []byte) (int, error) {
	w.write(string(bytes))
	return len(bytes), nil
}

func GetLogWriter(level string) (*LogWriter, error) {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	return &LogWriter{
		write: func(s string) {
			defaultLogger.Logf(l, s)
		}}, nil
}

func MustGetLogWriter(level string) *LogWriter {
	if w, err := GetLogWriter(level); err != nil {
		panic(err)
	} else {
		return w
	}
}
