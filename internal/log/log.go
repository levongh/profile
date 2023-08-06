package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	common "github.com/levongh/profile/common/config"
)

func NewLogger(service string, logLevel Level) (*Logger, error) {
	zapLevel := getZapLevel(logLevel)

	cfg := newZapConfig(service, zapLevel)
	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return newLogger(zl), nil
}

// New is deprecated, use NewLogger
// deprecated
func New(mode, serviceName string, opts ...Option) (*Logger, error) {
	zapLogger, err := newZap(mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create zap sugared: %w", err)
	}

	logger := &Logger{
		zapLogger: zapLogger.Sugar(),
	}
	for _, opt := range opts {
		opt(logger)
	}

	if logger.sentryOption.sentryDsn == "" || mode != common.ModeProd {
		return logger, nil
	}

	sentryCore, err := newSentryCore(newSentryOptions(logger.sentryOption.sentryDsn, mode, serviceName))
	if err != nil {
		return nil, fmt.Errorf("failed to init sentry core: %w", err)
	}

	core := zapcore.NewTee(zapLogger.Core(), sentryCore)
	logger.zapLogger = zap.New(core).Sugar()

	return logger, nil
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.zapLogger.Infow(msg, fieldsToInterface(fields)...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.zapLogger.Errorw(msg, fieldsToInterface(fields)...)
}

func (l *Logger) Debug(msg string, fields ...Field) {
	l.zapLogger.Debugw(msg, fieldsToInterface(fields)...)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.zapLogger.Warnw(msg, fieldsToInterface(fields)...)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.zapLogger.Infof(template, args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.zapLogger.Errorf(template, args...)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.zapLogger.Debugf(template, args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.zapLogger.Warnf(template, args...)
}

func (l *Logger) AddField(name, value string) *Logger {
	return &Logger{
		zapLogger:    l.zapLogger.With(name, value),
		sentryOption: l.sentryOption,
	}
}

func newZap(mode string) (*zap.Logger, error) {
	opts := []zap.Option{zap.AddCallerSkip(1)}

	if mode == modeDev {
		return zap.NewDevelopment(opts...)
	}

	return zap.NewProduction(opts...)
}

func newZapConfig(service string, level zapcore.Level) zap.Config {
	return zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "msg",
			LevelKey:       "lvl",
			TimeKey:        "ts",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochNanosTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]interface{}{
			"source": service,
		},
	}
}
