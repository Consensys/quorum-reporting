package log

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	ErrorLevel int = iota //0
	WarnLevel             //1
	InfoLevel             //2
	DebugLevel            //3
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, ForceColors: true})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)
}

func Error(args ...interface{}) {
	logMsg(logrus.ErrorLevel, args...)
}

func Warn(args ...interface{}) {
	logMsg(logrus.WarnLevel, args...)
}

func Info(args ...interface{}) {
	logMsg(logrus.InfoLevel, args...)
}

func Debug(args ...interface{}) {
	logMsg(logrus.DebugLevel, args...)
}

func Trace(args ...interface{}) {
	logMsg(logrus.TraceLevel, args...)
}

func logMsg(lvl logrus.Level, args ...interface{}) {
	if len(args) == 0 {
		return
	}

	// The first element is the main message, and all other arguments
	// must be in pairs, meaning there should be an odd number of
	// total args
	if len(args)%2 == 0 {
		args = append(args, "")
	}

	entries := make(map[string]interface{})
	for i := 1; i < len(args); i += 2 {
		entries[fmt.Sprintf("%v", args[i])] = args[i+1]
	}
	logrus.WithFields(entries).Log(lvl, args[0])
}
