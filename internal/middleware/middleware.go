package middleware

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
)

// HTTPLoggerMiddleware логирует информацию о запросе и ответе.
//
// Логирование выполняется после обработки запроса (после c.Next()):
//   - url, method, duration,
//   - status и size ответа.
//
// Middleware предполагает, что writer реализует gin.ResponseWriter.

// InitLogger создаёт production-логгер zap и возвращает SugaredLogger.
//
// При ошибке инициализации возвращает noop-логгер, чтобы сервис мог продолжить работу.
func InitLogger() *zap.SugaredLogger {
	cfg := zap.NewDevelopmentConfig()

	cfg.Encoding = "console"
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.DisableStacktrace = true

	logger, err := cfg.Build()
	if err != nil {
		log.Printf("failed to initialize zap logger: %v", err)
		return zap.NewNop().Sugar()
	}

	return logger.Sugar()
}
