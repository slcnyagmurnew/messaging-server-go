package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

// Logger is the core zap.Logger instance
var Logger *zap.Logger

// Sugar is the global SugaredLogger reference
var Sugar *zap.SugaredLogger

// Init initializes the global Logger and SugaredLogger.
func Init() {
	var err error
	var lvl zapcore.Level

	runEnv := os.Getenv("ENVIRONMENT")

	if runEnv == "DEVELOPMENT" {
		cfg := zap.NewDevelopmentConfig()
		lvl = zapcore.DebugLevel
		cfg.Level = zap.NewAtomicLevelAt(lvl)
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		Logger, err = cfg.Build()
	} else {
		cfg := zap.NewProductionConfig()

		// override log level if LOG_LEVEL is set
		if levelEnv := os.Getenv("LOG_LEVEL"); levelEnv != "" {
			if err := lvl.UnmarshalText([]byte(strings.ToLower(levelEnv))); err == nil {
				cfg.Level = zap.NewAtomicLevelAt(lvl)
			} else {
				fmt.Printf("invalid LOG_LEVEL %q, using default %q\n", levelEnv, cfg.Level.Level().String())
			}
		}

		Logger, err = cfg.Build()
	}
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	// Create SugaredLogger for ergonomic logging
	Sugar = Logger.Sugar()
}

// Sync flushes any buffered logs. Call this before your application exits.
func Sync() {
	_ = Logger.Sync()
}
