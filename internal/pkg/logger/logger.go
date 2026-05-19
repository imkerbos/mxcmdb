package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log 全局日志实例
var Log *zap.SugaredLogger

// Init 初始化 zap 日志
func Init(level, format string) error {
	var cfg zap.Config
	if format == "json" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	switch level {
	case "debug":
		cfg.Level.SetLevel(zapcore.DebugLevel)
	case "info":
		cfg.Level.SetLevel(zapcore.InfoLevel)
	case "warn":
		cfg.Level.SetLevel(zapcore.WarnLevel)
	case "error":
		cfg.Level.SetLevel(zapcore.ErrorLevel)
	default:
		cfg.Level.SetLevel(zapcore.InfoLevel)
	}

	l, err := cfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		return fmt.Errorf("build logger: %w", err)
	}
	Log = l.Sugar()
	return nil
}
