package globals

import (
	"context"
	"hitman/api/v1alpha1"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	ExecContext = ExecutionContext{
		Context: context.Background(),
	}
)

// ExecutionContext TODO
type ExecutionContext struct {
	Context context.Context
	Logger  zap.SugaredLogger
	Config  v1alpha1.ConfigT

	//
	LogLevel string
	DryRun   bool
}

// SetLogger TODO
func SetLogger(logLevel string, disableTrace bool) (err error) {
	parsedLogLevel, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		return err
	}

	// Initialize the logger
	loggerConfig := zap.NewProductionConfig()
	if disableTrace {
		loggerConfig.DisableStacktrace = true
		loggerConfig.DisableCaller = true
	}

	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	loggerConfig.Level.SetLevel(parsedLogLevel.Level())

	// Configure the logger
	logger, err := loggerConfig.Build()
	if err != nil {
		return err
	}

	ExecContext.Logger = *logger.Sugar()
	return nil
}
