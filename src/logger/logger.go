package logger

import "go.uber.org/zap"

var (
	logger *zap.Logger
	err    error
)

// Initializes a new zap logger with default settings unless specified otherwise
func New(c *zap.Config, prod bool) (*zap.Logger, error) {
	if prod {
		return zap.NewProductionConfig().Build()
	}

	return zap.NewDevelopmentConfig().Build()
}
