package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	Debug Level = "debug"
	Info  Level = "info"
	Warn  Level = "warn"
	Err   Level = "error"
)

var (
	levelToZap = map[Level]zapcore.Level{
		Debug: zap.DebugLevel,
		Info:  zap.InfoLevel,
		Warn:  zap.WarnLevel,
		Err:   zap.ErrorLevel,
	}
	defaultLevel = zap.InfoLevel
)

type Level string

type Option func(*Logger)

type Logger struct {
	zapLogger    *zap.SugaredLogger
	sentryOption sentryOption
}

func newLogger(zap *zap.Logger) *Logger {
	return &Logger{
		zapLogger: zap.Sugar(),
	}
}

// NewTestLogger return instance of Logger that discards all output.
func NewTestLogger() *Logger {
	return &Logger{
		zapLogger: zap.NewNop().Sugar(),
	}
}

func getZapLevel(l Level) zapcore.Level {
	out, ok := levelToZap[l]
	if !ok {
		return defaultLevel
	}
	return out
}
