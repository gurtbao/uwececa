package config

import (
	"context"

	envconfig "github.com/sethvargo/go-envconfig"
)

type Config struct {
	Core   Core   `env:",prefix=UWECECA_"`
	DB     DB     `env:",prefix=UWECECA_DB_"`
	Mailer Mailer `env:",prefix=UWECECA_MAILER_"`
}

type Core struct {
	Development bool   `env:"DEVELOPMENT,default=false"`
	Addr        string `env:"ADDR,default=localhost:3000"`
	BaseDomain  string `env:"DOMAIN,default=localhost:3000"`
}

type DB struct {
	Location string `env:"LOCATION,default=db.sqlite3"`
}

type Mailer struct {
	Host        string `env:"HOST,default=localhost"`
	Port        int    `env:"PORT,default=1025"`
	Username    string `env:"USERNAME,required"`
	Password    string `env:"PASSWORD,required"`
	FromAddress string `env:"FROM_ADDR,required"`
}

func Load(ctx context.Context) (*Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
