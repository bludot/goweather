package logger

import "go.uber.org/zap"

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger() *ZapLogger {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	return &ZapLogger{logger}
}
