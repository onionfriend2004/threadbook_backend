package main

import (
	goLog "log"
	"os"

	"github.com/onionfriend2004/threadbook_backend/config"
	"github.com/onionfriend2004/threadbook_backend/internal/app"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/logger"
	"go.uber.org/zap"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		goLog.Fatalf("failed to load config: %v", err)
	}

	// Создание Логера
	log := logger.New(cfg)
	defer log.Sync()

	log.Info("Hello world!")

	// Запуск приложения
	if err := app.Run(cfg, log); err != nil {
		log.Error("application failed", zap.Error(err))
		os.Exit(1)
	}
}
