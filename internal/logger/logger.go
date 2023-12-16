package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func init() {
	// Create a logger
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	Log, _ = config.Build()
}

const (
	green  = "\033[32m"
	blue   = "\033[34m"
	yellow = "\033[33m"
	purple = "\033[35m"
	reset  = "\033[0m"
)

func GreensInfo(msg string, args ...any) {
	Log.Info(green + fmt.Sprintf(msg, args...) + reset)
}

func GreensError(msg string, args ...any) {
	Log.Error(green + fmt.Sprintf(msg, args...) + reset)
}

func RodentInfo(msg string, args ...any) {
	Log.Info(yellow + fmt.Sprintf(msg, args...) + reset)
}

func RodentError(msg string, args ...any) {
	Log.Error(yellow + fmt.Sprintf(msg, args...) + reset)
}

func KernelInfo(msg string, args ...any) {
	Log.Info(purple + fmt.Sprintf(msg, args...) + reset)
}

func BrokerInfo(isOn bool, msg string, args ...any) {
	if isOn {
		Log.Info(blue + fmt.Sprintf(msg, args...) + reset)
	}
}
