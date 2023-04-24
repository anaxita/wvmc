package app

import (
	"go.uber.org/zap"
)

func NewLogger(logFiles ...string) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	cfg.OutputPaths = append(cfg.OutputPaths, logFiles...)

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
