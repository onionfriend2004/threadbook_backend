package infra

import (
	"fmt"

	"github.com/onionfriend2004/threadbook_backend/config"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Подключение к SQL-Database (PostgreSQL)
func PostgresConnect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Postgres.Host,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Name,
		cfg.Postgres.Port,
		cfg.Postgres.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Миграции
	err = db.AutoMigrate(&domain.User{})
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// db.Exec(``) Кастомные запросы DDL

	return db, nil
}
