package config

import (
	"github.com/spf13/viper"
)

// Config содержит конфигурацию приложения.
type Config struct {
	Postgres struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"database"`

	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"` // номер базы (по умолчанию 0)
	} `mapstructure:"redis"`

	Log struct {
		Level string `mapstructure:"level"` // e.g. "debug", "info"
	} `mapstructure:"log"`
}

// LoadConfig загружает конфигурацию из файла YAML и переменных среды.
func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// If allow env vars like LOG_LEVEL, DB_HOST, etc.
	// viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Установка разумных значений (дефолтов) по умолчанию
	viper.SetDefault("log.level", "info")

	// Чтение конфига
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Запись в конфиг
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
