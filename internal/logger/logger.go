package logger

import "go.uber.org/zap"

var Logger *zap.SugaredLogger

func init() {
	logger, _ := zap.NewProduction()
	Logger = logger.Sugar()
}
