package logging

import "go.uber.org/zap"

func GetLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	return logger
}
