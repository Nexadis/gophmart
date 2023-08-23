package logger

import "go.uber.org/zap"

var Logger *zap.SugaredLogger

func init() {
	config := zap.NewProductionConfig()
	config.Level.SetLevel(zap.DebugLevel)
	logger, _ := config.Build()
	Logger = logger.Sugar()
}
