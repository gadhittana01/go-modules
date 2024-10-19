package utils

import (
	"database/sql"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	var err error
	config := zap.NewProductionConfig()
	logger, err = config.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

func LogIfError(err error, message ...string) {
	if err != nil {
		if err == pgx.ErrNoRows || err == sql.ErrNoRows {
			if len(message) > 0 && message[0] != "" {
				LogWarning(message[0], zap.Error(err))
			} else {
				LogWarning("error occurred", zap.Error(err))
			}
		} else {
			if len(message) > 0 && message[0] != "" {
				LogError(message[0], zap.Error(err))
			} else {
				LogError("error occurred", zap.Error(err))
			}
		}
	}
}

func LogInfo(message string, fields ...zap.Field) {
	logger.Info(message, fields...)
}

func LogFatal(message string, fields ...zap.Field) {
	logger.Fatal(message, fields...)
}

func LogPanic(message string, fields ...zap.Field) {
	logger.Panic(message, fields...)
}

func LogDebug(message string, fields ...zap.Field) {
	logger.Debug(message, fields...)
}

func LogError(message string, fields ...zap.Field) {
	logger.Error(message, fields...)
}

func LogWarning(message string, fields ...zap.Field) {
	logger.Warn(message, fields...)
}
