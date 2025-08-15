package config

import (
	"context"

	envconfig "github.com/sethvargo/go-envconfig"
)

type Config struct {
	Core Core `env:",prefix=UWECECA_"`
	DB   DB   `env:",prefix=UWECECA_DB_"`
}

type Core struct {
	Development bool   `env:"DEVELOPMENT,default=false"`
	Addr        string `env:"ADDR,default=localhost:3000"`
}

type DB struct {
	Location string `env:"LOCATION,default=db.sqlite3"`
}

func Load(ctx context.Context) (*Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
