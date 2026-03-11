package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"telephony/pkg/logger"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Env string

var (
	Local Env = "local"
	Dev   Env = "dev"
	Prod  Env = "prod"
)

type Config struct {
	*jsonConfig
	*envConfig
}

type envConfig struct {
	Env       Env `env:"ENV" env-required:"true"`
	Postgres  PostgresConfig
	Redis     RedisConfig
	Auth      AuthConfig
	Storage   StorageConfig
	EmailEnv  SMSConfig
	S3        StorageConfig
	Sms       SMSConfig
	Payment   PaymentConfig
	Telephony TelephonyConfig
}

type PaymentConfig struct {
	Robokassa RobokassaConfig
}

type RobokassaConfig struct {
	MerchantLogin string `env:"ROBOKASSA_MERCHANT_LOGIN"`
	Password1     string `env:"ROBOKASSA_PASSWORD_1"`
	Password2     string `env:"ROBOKASSA_PASSWORD_2"`
	TestMode      bool   `env:"ROBOKASSA_TEST_MODE"`
}

// TelephonyConfig описывает конфигурацию телефонии и провайдеров.
type TelephonyConfig struct {
	// Enabled провайдеры, управляемые через переменные окружения.
	// Примеры:
	// TELEPHONY_TWILIO_ENABLED=true
	// TELEPHONY_MANGO_ENABLED=true
	// TELEPHONY_MTS_ENABLED=false
	// TELEPHONY_ZADARMA_ENABLED=true
	TwilioEnabled  bool `env:"TELEPHONY_TWILIO_ENABLED" env-default:"false"`
	MangoEnabled   bool `env:"TELEPHONY_MANGO_ENABLED" env-default:"false"`
	MTSEnabled     bool `env:"TELEPHONY_MTS_ENABLED" env-default:"false"`
	ZadarmaEnabled bool `env:"TELEPHONY_ZADARMA_ENABLED" env-default:"false"`

	// Секреты/ключи для проверки подписей вебхуков (если используются у провайдера).
	TwilioWebhookSecret  string `env:"TELEPHONY_TWILIO_WEBHOOK_SECRET"`
	MangoWebhookSecret   string `env:"TELEPHONY_MANGO_WEBHOOK_SECRET"`
	MTSWebhookSecret     string `env:"TELEPHONY_MTS_WEBHOOK_SECRET"`
	ZadarmaWebhookSecret string `env:"TELEPHONY_ZADARMA_WEBHOOK_SECRET"`
}

type jsonConfig struct {
	Server  ServerConfig  `json:"server" mapstructure:"server" validate:"required"`
	Handler HandlerConfig `json:"handler" mapstructure:"handler" validate:"required"`
}

type ServerConfig struct {
	Port           int           `json:"port" mapstructure:"port" validate:"required"`
	ReadTimeout    time.Duration `json:"read_timeout" mapstructure:"read_timeout" validate:"required"`
	WriteTimeout   time.Duration `json:"write_timeout" mapstructure:"write_timeout" validate:"required"`
	MaxHeaderBytes int           `json:"max_header_bytes" mapstructure:"max_header_bytes" validate:"required"`
}

type HandlerConfig struct {
	RequestTimeout     time.Duration `json:"request_timeout" mapstructure:"request_timeout" validate:"required"`
	AllowedCORSOrigins string        `json:"allowed_cors_origins" mapstructure:"allowed_cors_origins" validate:"required"` // comma-separated origins or "*"
}

type PostgresConfig struct {
	Host           string        `env:"POSTGRES_HOST" env-required:"true"`
	Port           int           `env:"POSTGRES_PORT" env-required:"true"`
	User           string        `env:"POSTGRES_USER" env-required:"true"`
	Password       string        `env:"POSTGRES_PASSWORD" env-required:"true"`
	Database       string        `env:"POSTGRES_DATABASE" env-required:"true"`
	IsDebug        bool          `env:"POSTGRES_IS_DEBUG" env-default:"false"`
	RequestTimeout time.Duration `env:"POSTGRES_REQUEST_TIMEOUT" env-required:"true"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" env-required:"true"`
	Port     string `env:"REDIS_PORT" env-required:"true"`
	Password string `env:"REDIS_PASSWORD" env-required:"true"`
	DB       int    `env:"REDIS_DB" env-required:"true"`
}

func MustConfig(log logger.Logger) *Config {
	// Load `.env` from the current dir or any parent dir (useful for tests executed from subpackages).
	if err := loadDotEnvUpwards(); err != nil {
		// Keep behavior consistent: config requires env vars to be present.
		log.Panic(fmt.Sprintf("failed download the file .env: %v", err))
	}

	path := fetchConfigPath()
	if path == "" {
		log.Panic("config path is empty")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Panic("config file does not exist: " + path)
	}

	viper.SetConfigFile(path)
	viper.SetConfigType("json")

	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("Error reading config file: %v", err)
	}

	var jsonCfg jsonConfig
	if err := viper.Unmarshal(&jsonCfg, viper.DecodeHook(
		mapstructure.StringToTimeDurationHookFunc(),
	)); err != nil {
		log.Panicf("unable to decode into struct: %v", err)
	}

	validate := validator.New()
	if err := validate.Struct(jsonCfg); err != nil {
		log.Panicf("unable to validate config file: %v", err)
	}
	var envCfg envConfig

	err := cleanenv.ReadEnv(&envCfg)
	if err != nil {
		log.Panic("failed to read envConfig: " + err.Error())
	}
	return &Config{
		jsonConfig: &jsonCfg,
		envConfig:  &envCfg,
	}
}

func loadDotEnvUpwards() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	dir := wd
	for i := 0; i < 10; i++ {
		envPath := filepath.Join(dir, ".env")
		if _, statErr := os.Stat(envPath); statErr == nil {
			return godotenv.Load(envPath)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback: try default behavior (may rely on process working directory).
	return godotenv.Load()
}

func fetchConfigPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()
	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}
	return res
}
