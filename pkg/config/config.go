package config

import (
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const (
	envLocalFileName      = ".env.local"
	envFileName           = ".env"
	envProductionFileName = ".env.production"
)

type ServerParams struct {
	Handshake time.Duration `mapstructure:"handshake" validate:"required"`
	PoW       struct {
		Difficulty   uint          `mapstructure:"difficulty" validate:"required"`
		SaltLen      uint          `mapstructure:"salt_len" validate:"required"`
		ChallengeTTL time.Duration `mapstructure:"challenge_ttl" validate:"required"`
	}
}

type ClientParams struct {
	PoW struct {
		SaltLen uint `mapstructure:"salt_len" validate:"required"`
	}
}

type Config[T any] struct {
	App struct {
		Addr string `mapstructure:"addr" validate:"required,hostname_port"`
		Name string `mapstructure:"name" validate:"required"`
	}
	TLS struct {
		Enabled bool   `mapstructure:"enabled"`
		Cert    string `mapstructure:"cert" validate:"required"`
		Key     string `mapstructure:"key" validate:"required"`
	}
	Log struct {
		Level string `mapstructure:"level" validate:"required"`
	}
	Params T `mapstructure:"params" validate:"required"`
}

func New[T any](filePath string) *Config[T] {
	_ = godotenv.Load(envLocalFileName, envFileName, envProductionFileName)

	v := viper.New()
	v.SetConfigFile(filePath)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.SetEnvPrefix("")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	cfg := &Config[T]{}
	if err := v.Unmarshal(cfg); err != nil {
		panic(err)
	}

	if err := validator.New().Struct(cfg); err != nil {
		panic(err)
	}

	return cfg
}
