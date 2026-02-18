package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a configured zap logger.
// In production mode it uses JSON encoding; in development — console with colors.
func New(env string) (*zap.Logger, error) {
	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// Fatal logs a message at Fatal level and calls os.Exit(1).
// Used only in main before the application fully starts.
func Fatal(msg string, err error) {
	l, _ := zap.NewProduction()
	l.Fatal(msg, zap.Error(err))
	os.Exit(1)
}
