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
