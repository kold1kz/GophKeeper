// Package middleware содержит вспомогательные middleware-компоненты
// и инфраструктурные элементы приложения, связанные с логированием.
package middleware

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
)

// InitLogger создаёт и настраивает логгер zap.
//
// Логгер создаётся в человекочитаемом консольном формате,
// с уровнем логирования Info и цветным отображением уровней.
//
// Если инициализация логгера завершается ошибкой,
// функция возвращает noop-логгер, чтобы приложение
// могло продолжить работу.
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
