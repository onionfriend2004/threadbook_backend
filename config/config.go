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

	Nats struct {
		Host              string `mapstructure:"host"`
		Port              int    `mapstructure:"port"`
		Name              string `mapstructure:"name"`                // имя подключения для мониторинга и логов
		Timeout           int    `mapstructure:"timeout"`             // таймаут подключения в секундах
		MaxReconnects     int    `mapstructure:"max_reconnects"`      // -1 = бесконечно
		ReconnectWait     int    `mapstructure:"reconnect_wait_ms"`   // задержка между попытками в миллисекундах
		VerifyCodeSubject string `mapstructure:"verify_code_subject"` // топик для отправки событий верификации кода (куда срать)
	} `mapstructure:"nats"`

	Argon2 struct {
		Memory      uint32 `mapstructure:"memory"`      // память в КБ (например, 64*1024 = 64 МБ)
		Iterations  uint32 `mapstructure:"iterations"`  // количество итераций (2 Dev, 3 Prod)
		Parallelism uint8  `mapstructure:"parallelism"` // количество параллельных потоков (число ядер CPU)
		SaltLength  uint32 `mapstructure:"salt_length"` // длина соли в байтах (обычно 16)
		KeyLength   uint32 `mapstructure:"key_length"`  // длина хэша в байтах (обычно 32)
	} `mapstructure:"argon2"`

	Minio struct {
		Host      string `mapstructure:"host"`
		Port      int    `mapstructure:"port"`
		AccessKey string `mapstructure:"access_key"`
		SecretKey string `mapstructure:"secret_key"`
		UseSSL    bool   `mapstructure:"use_ssl"`
		Bucket    string `mapstructure:"bucket"` // можно для основного бакета
	} `mapstructure:"minio"`

	Centrifugo struct {
		Host        string `mapstructure:"host"`
		Port        int    `mapstructure:"port"`
		UseSSL      bool   `mapstructure:"use_ssl"`
		APIKey      string `mapstructure:"api_key"`       // ключ для админ/серверных запросов
		TokenHMAC   string `mapstructure:"token_hmac"`    // секрет для подписания user токенов (JWT)
		AdminAPIKey string `mapstructure:"admin_api_key"` // ключ для административного API (если нужно)
		Namespace   string `mapstructure:"namespace"`     // например, "chat"
		TTL         uint32 `mapstructure:"ttl"`           // TTL токена подключения пользователя в секундах
	} `mapstructure:"centrifugo"`

	UserSession struct {
		TTL uint32 `mapstructure:"ttl"` // TTL Жизни сессии пользователя
	} `mapstructure:"user_session"`

	VerifyCode struct {
		TTL uint32 `mapstructure:"ttl"` // TTL Жизни кода для подтверждения почты
	} `mapstructure:"verify_code"`

	Cookie struct {
		SessionCookieName string `mapstructure:"session_cookie_name"` // Рекомендуется использовать нейтральное имя (например, "sid"), чтобы не раскрывать детали реализации.
		SessionCookiePath string `mapstructure:"session_cookie_path"` // Путь, для которого устанавливается кука.  Обычно "/" — чтобы кука была доступна всем эндпоинтам API.
		SessionDomain     string `mapstructure:"session_domain"`      //Домен, для которого действует кука. Оставьте пустым если нет поддоменов
		SessionSecure     bool   `mapstructure:"session_secure"`      // HTTPS. Обязательно true в production! В development (HTTP) должно быть false.
		SessionSameSite   string `mapstructure:"session_samesite"`    // От CSFR Защита
		// - "Strict": кука не отправляется при переходе с внешних сайтов (макс. безопасность).		<- в production Strict + https
		// - "Lax": разрешает отправку при безопасных GET-запросах (например, клик по ссылке).		<- в dev можно Lax + http
		// - "None": отключает защиту (требует SessionSecure=true).									<- не надо дядя
	} `mapstructure:"cookie"`

	Smtp struct {
		Server   string `mapstructure:"server"`
		Port     string `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Sender   string `mapstructure:"sender"`
	} `mapstructure:"smtp"`

	LiveKit struct {
		URL       string `mapstructure:"url"`
		APIKey    string `mapstructure:"api_key"`
		APISecret string `mapstructure:"api_secret"`
	} `mapstructure:"livekit"`

	Room struct {
		EmptyTTL        uint32 `mapstructure:"empty_ttl"`        // В секундах
		MaxParticipants uint32 `mapstructure:"max_participants"` // в штуках)
	} `mapstructure:"room"`

	Log struct {
		Level string `mapstructure:"level"` // e.g. "debug", "info"
	} `mapstructure:"log"`

	Upload struct {
		Common struct {
			AllowedFormats []string `mapstructure:"allowed_formats"` // глобально разрешённые форматы (png, jpg, webp, mp4 и т.д.)
			MaxSizeMB      int      `mapstructure:"max_size_mb"`     // общий лимит (на всякий случай)
		} `mapstructure:"common"`

		Spool struct {
			MaxBannerSizeMB int `mapstructure:"max_banner_size_mb"`
		} `mapstructure:"spool"`
	}
}

// LoadConfig загружает конфигурацию из файла YAML и переменных среды.
func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml") // or "env"

	// If allow env vars like LOG_LEVEL, DB_HOST, etc.
	// viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) это говно понадопится если env юзать
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
