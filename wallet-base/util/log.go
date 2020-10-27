package util

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

// InitDefaultRotationLogger init rotation log config with 7 days maxAge and json format.
func InitDefaultRotationLogger(filePath, fileName string) error {
	return InitDaysJSONRotationLogger(filePath, fileName, 7)
}

// InitDaysJSONRotationLogger init rotation log config with maxAgeDays and json format.
func InitDaysJSONRotationLogger(filePath, fileName string, maxAgeDays uint) error {
	const day = time.Hour * 24
	return InitRotationLogger(filePath, fileName, time.Duration(maxAgeDays)*day, day, &logrus.JSONFormatter{})
}

// InitRotationLogger init rotation log config.
func InitRotationLogger(filePath, fileName string, maxAge, rotationTime time.Duration, formatter logrus.Formatter) error {
	err := os.MkdirAll(filePath, 0700)
	if err != nil {
		return err
	}

	filePath, err = filepath.Abs(filePath + fileName)
	if err != nil {
		return err
	}

	writer, err := rotatelogs.New(
		filePath+".%Y%m%d%H%M%S",
		rotatelogs.WithLinkName(filePath),
		rotatelogs.WithMaxAge(maxAge),
		rotatelogs.WithRotationTime(rotationTime),
	)
	if err != nil {
		return err
	}

	logrus.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			logrus.DebugLevel: writer,
			logrus.InfoLevel:  writer,
			logrus.WarnLevel:  writer,
			logrus.ErrorLevel: writer,
			logrus.FatalLevel: writer,
			logrus.PanicLevel: writer,
		},
		formatter,
	))
	return nil
}

type LogWriter struct {
	logFunc func(...interface{})
}

func NewLogWriter(logFunc func(...interface{})) *LogWriter {
	return &LogWriter{
		logFunc: logFunc,
	}
}

func (l *LogWriter) Write(p []byte) (n int, err error) {
	if l.logFunc != nil {
		l.logFunc(string(p))
	}

	return len(p), nil
}

func TraceProgress(tag string, num, step, total int) {
	if num == total {
		fmt.Printf("\r%s: 100.0%%", tag)
		fmt.Println()
		return
	}

	if num%step == 0 {
		fmt.Printf("\r%s: %.1f%%", tag, 100*float64(num)/float64(total))
	}
}
