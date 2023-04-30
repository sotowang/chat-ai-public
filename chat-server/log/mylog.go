package mylog

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"path"
	"runtime"

	"time"
)

var Logger = &logrus.Entry{}

func InitLog() {
	file := "./shiyu.log"
	writer, err := rotatelogs.New(
		file+"-%Y%m%d",
		rotatelogs.WithLinkName(file),
		rotatelogs.WithMaxAge(time.Duration(3*24)*time.Hour),
	)
	if err != nil {
		logrus.Info("打开日志文件失败，默认输出到stderr")
	} else {
		// 将日志输出到标准输出，就是直接在控制台打印出来。
		logrus.SetOutput(writer)
	}
	logrus.SetLevel(logrus.InfoLevel)
	// 设置为true则显示日志在代码什么位置打印的
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(frame *runtime.Frame) (function string, f string) {
			filename := fmt.Sprintf("%s l:%d", path.Base(frame.File), frame.Line)
			return "", filename
		},
		QuoteEmptyFields: false,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:        "t",
			logrus.FieldKeyLevel:       "l",
			logrus.FieldKeyMsg:         "m",
			logrus.FieldKeyFile:        "f",
			logrus.FieldKeyLogrusError: "err",
		},
	})
}
