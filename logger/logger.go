package logger

import (
	"go.uber.org/zap"
)

func NewLogger(isProduction bool) (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	if isProduction {
		logger, err = zap.NewProduction()
		if err != nil {
			return nil, err
		}
	}

	return logger, nil
}
