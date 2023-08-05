package config

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/levongh/profile/common"
	"github.com/levongh/profile/internal/log"
)

type Config struct {
	Port        string    `envconfig:"PORT" validate:"required,startswith=:"`
	Host        string    `envconfig:"HOST" validate:"required,uri"`
	ClientHost  string    `envconfig:"CLIENT_HOST" validate:"required,uri"`
	Mode        string    `envconfig:"MODE" validate:"required,oneof='local' 'development' 'staging' 'production'"`
	ServiceName string    `envconfig:"SERVICE_NAME" validate:"required"`
	LogLevel    log.Level `envconfig:"LOG_LEVEL"`
	StorageDSN  string    `envconfig:"STORAGE_DSN" validate:"required,uri"`

	InternalAPIUser     string `envconfig:"INTERNAL_API_USER" validate:"required"`
	InternalAPIPassword string `envconfig:"INTERNAL_API_PASSWORD" validate:"required"`
}

func Read() (*Config, error) {
	_ = godotenv.Overload(".env", ".env.local")
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c Config) HostWithoutProtocol() string {
	wo := strings.TrimPrefix(c.Host, "http://")
	return strings.TrimPrefix(wo, "https://")
}

func (c Config) IsNoop() bool {
	return c.Mode == common.ModeDev
}
