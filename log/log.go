package log

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/where"
	"os"
	"path/filepath"
	"time"
)

var writeLogs bool

func Setup() error {
	writeLogs = viper.GetBool(key.LogsWrite)

	if !writeLogs {
		return nil
	}

	logsPath := where.Logs()

	if logsPath == "" {
		return errors.New("logs path is not set")
	}

	today := time.Now().Format("2006-01-02")
	logFilePath := filepath.Join(logsPath, fmt.Sprintf("%s.log", today))
	logFile, err := filesystem.Api().OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o640)
	if err != nil {
		return err
	}

	log.SetOutput(logFile)

	if viper.GetBool(key.LogsJson) {
		log.SetFormatter(&log.JSONFormatter{PrettyPrint: true})
	} else {
		log.SetFormatter(&log.TextFormatter{})
	}

	switch viper.GetString(key.LogsLevel) {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	return nil
}

func Error(args ...interface{}) {
	if writeLogs {
		log.Error(args...)
	}
}

func Warn(args ...interface{}) {
	if writeLogs {
		log.Warn(args...)
	}
}

func Warnf(format string, args ...interface{}) {
	if writeLogs {
		log.Warnf(format, args...)
	}
}

func Info(args ...interface{}) {
	if writeLogs {
		log.Info(args...)
	}
}

func Infof(format string, args ...interface{}) {
	if writeLogs {
		log.Infof(format, args...)
	}
}

func Debugf(format string, args ...interface{}) {
	if writeLogs {
		log.Debugf(format, args...)
	}
}

func Tracef(format string, args ...interface{}) {
	if writeLogs {
		log.Tracef(format, args...)
	}
}
