package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port            int           `env:"PORT,required"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT,required"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT,required"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT,required"`
	PostgresUri     string        `env:"POSTGRES_URI,required"`
	JwtAccessSecret string        `env:"JWT_ACCESS_SECRET,required"`
	JwtAccessExpire time.Duration `env:"JWT_ACCESS_EXPIRE,required"`
	SessionExpire   time.Duration `env:"SESSION_EXPIRE,required"`
	GoEnv           string        `env:"GO_ENV,required"`
	TrustedOrigin   string        `env:"TRUSTED_ORIGIN,required"`
}

func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
