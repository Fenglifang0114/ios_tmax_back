package log

import (
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

var (
	Log  *log.Logger
	once sync.Once
)

func init() {
	// once ensures the singleton is initialized only once
	once.Do(func() {
		log := &log.Logger{
			Out:   os.Stderr,
			Level: log.DebugLevel, //DebugLevel WarnLevel调试LOG
			Formatter: &log.TextFormatter{
				ForceColors:     true,
				FullTimestamp:   true,
				TimestampFormat: "15:04:05",
				FieldMap: log.FieldMap{
					"FieldKeyTime":  "@timestamp",
					"FieldKeyLevel": "@level",
					"FieldKeyMsg":   "@message",
				},
			},
		}
		log.SetReportCaller(true)
		Log = log
	})
}

//输出log

// func init() {
// 	// once ensures the singleton is initialized only once
// 	once.Do(func() {
// 		// 打开一个文件用于写入日志，如果文件不存在则创建，如果存在则追加内容
// 		logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
// 		if err != nil {
// 			log.Fatalf("Failed to open log file: %v", err)
// 		}

// 		log := &log.Logger{
// 			// 将日志输出到文件
// 			Out:   logFile,
// 			Level: log.DebugLevel, //DebugLevel WarnLevel调试LOG
// 			Formatter: &log.TextFormatter{
// 				ForceColors:     true, // 文件中不需要颜色
// 				FullTimestamp:   true,
// 				TimestampFormat: "15:04:05",
// 				FieldMap: log.FieldMap{
// 					"FieldKeyTime":  "@timestamp",
// 					"FieldKeyLevel": "@level",
// 					"FieldKeyMsg":   "@message",
// 				},
// 			},
// 		}
// 		log.SetReportCaller(true)
// 		Log = log
// 	})
// }
