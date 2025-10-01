package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TODO: Миграции, декомпозиция, ORM и тп, ReNames
func GetPostgres() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Получаем underlying sql.DB для проверки подключения
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Проверяем подключение
	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
