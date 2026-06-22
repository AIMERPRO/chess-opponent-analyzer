package logger

import "go.uber.org/zap"

func NewLogger(appEnv string) (*zap.Logger, error) {
	if appEnv == "development" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
