package logger

import (
	"git.sr.ht/~mcldresner/tfdog/config"
	"git.sr.ht/~mcldresner/tfdog/version"
	"go.uber.org/zap"
)

func NewLogger(cfg *config.Logger) (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	if cfg.IsProduction {
		logger, err = zap.NewProduction()
		if err != nil {
			return nil, err
		}
	}

	logger = logger.With(zap.String("version", version.Version))

	return logger, nil
}
