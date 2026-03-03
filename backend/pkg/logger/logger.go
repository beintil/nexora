package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	*zap.SugaredLogger
}

func NewLogger() (Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &logger{
		zapLogger.Sugar(),
	}, nil
}

func (l *logger) GetSugar() *zap.SugaredLogger {
	return l.SugaredLogger
}

func (l *logger) Warning(args ...interface{}) {
	l.SugaredLogger.Warn(args...)
}

func (l *logger) Warningf(format string, args ...interface{}) {
	l.SugaredLogger.Warnf(format, args...)
}

func (l *logger) Warningln(args ...interface{}) {
	l.SugaredLogger.Warnln(args...)
}

func (l *logger) Sync() {
	l.SugaredLogger.Sync()
}
