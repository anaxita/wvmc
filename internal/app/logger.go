package app

import (
	"fmt"
	"path"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(logFiles ...string) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()

	if len(logFiles) != 0 {
		dir, file := path.Split(logFiles[0])
		prefix := time.Now().Format("2006-01-02_15-04-05")
		cfg.OutputPaths = append(cfg.OutputPaths, fmt.Sprintf("%s/%s_%s", dir, prefix, file))
	}

	cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.TimeKey = "datetime"

	l, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return l.Sugar(), nil
}
