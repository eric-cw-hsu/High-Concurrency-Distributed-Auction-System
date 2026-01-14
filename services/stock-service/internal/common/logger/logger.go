package logger

import (
	"strings"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init initializes the global logger with provided config
func Init(cfg *config.LoggerConfig) *zap.Logger {
	var logger *zap.Logger
	var err error

	if cfg.Format == "json" {
		// Production: JSON format
		zapCfg := zap.NewProductionConfig()
		zapCfg.Level = zap.NewAtomicLevelAt(parseLevel(cfg.Level))
		zapCfg.EncoderConfig.TimeKey = "timestamp"
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapCfg.OutputPaths = []string{"stderr"}
		zapCfg.ErrorOutputPaths = []string{"stderr"}

		opts := []zap.Option{
			zap.AddCaller(),
		}
		if cfg.AddSource {
			opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
		}

		logger, err = zapCfg.Build(opts...)
	} else {
		// Development: Console format with colors
		zapCfg := zap.NewDevelopmentConfig()
		zapCfg.Level = zap.NewAtomicLevelAt(parseLevel(cfg.Level))
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapCfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		zapCfg.OutputPaths = []string{"stderr"}
		zapCfg.ErrorOutputPaths = []string{"stderr"}

		logger, err = zapCfg.Build(zap.AddCaller())
	}

	if err != nil {
		panic(err)
	}

	// Set as global logger
	zap.ReplaceGlobals(logger)

	logger.Info("logger initialized",
		zap.String("environment", cfg.Environment),
		zap.String("level", cfg.Level),
		zap.String("format", cfg.Format),
	)

	return logger
}

// parseLevel parses string level to zapcore.Level
func parseLevel(level string) zapcore.Level {
	level = strings.ToLower(level)

	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
