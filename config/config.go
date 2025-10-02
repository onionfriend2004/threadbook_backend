package config

import (
	"github.com/spf13/viper"
)

// Config содержит конфигурацию приложения.
// Короче, мне привычно писать конфиги в yml потому что у них есть уровни вложенности, а в env их нет,
// я хочу разграничивать логику, мне удобнее писать Redis.port и Postgres.port а не городить шляпу рядом с ними
type Config struct {
	App struct {
		Port string `mapstructure:"port"` // Порт размещения
	} `mapstructure:"app"`

	Postgres struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"` // слонярыч postgres (Шиша может другое имя дать, хз посмотрим)
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		SSLMode  string `mapstructure:"sslmode"` // подумать над безопасностью этого параметра, хз
	} `mapstructure:"database"`

	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"` // номер базы (по умолчанию 0)
	} `mapstructure:"redis"`

	Argon2 struct {
		Memory      uint32 `mapstructure:"memory"`      // память в КБ (например, 64*1024 = 64 МБ)
		Iterations  uint32 `mapstructure:"iterations"`  // количество итераций (2 Dev, 3 Prod)
		Parallelism uint8  `mapstructure:"parallelism"` // количество параллельных потоков (число ядер CPU)
		SaltLength  uint32 `mapstructure:"salt_length"` // длина соли в байтах (обычно 16)
		KeyLength   uint32 `mapstructure:"key_length"`  // длина хэша в байтах (обычно 32)
	} `mapstructure:"argon2"`

	User_session struct {
		TTL uint32 `mapstructure:"ttl"` // TTL Жизни сессии пользователя
	} `mapstructure:"user_session"`

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
