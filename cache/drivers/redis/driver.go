package redis

import "github.com/LouYuanbo1/go-webservice/cache"

type Driver struct {
	cfg *Config
}

// New 返回一个配置好的 Driver
func New(cfg *Config) *Driver {
	if cfg == nil {
		return &Driver{
			cfg: &Config{
				Host:          "localhost",
				Port:          6379,
				Password:      "",
				DB:            0,
				Protocol:      0,
				UnstableResp3: false,
			},
		}
	}
	return &Driver{
		cfg: &Config{
			Host:          cfg.Host,
			Port:          cfg.Port,
			Password:      cfg.Password,
			DB:            cfg.DB,
			Protocol:      cfg.Protocol,
			UnstableResp3: cfg.UnstableResp3,
		},
	}
}

func (d *Driver) Name() string {
	return "redis"
}

func (d *Driver) Initialize() (cache.Cache, error) {
	return newRedisCache(d.cfg)
}
