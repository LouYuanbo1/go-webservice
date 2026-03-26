package redis

import (
	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/singleflightx"
)

type Driver struct {
	cfg *Config
	sf  singleflightx.SingleFlight
}

// New 返回一个配置好的 Driver
func NewDriver(cfg *Config, sf singleflightx.SingleFlight) *Driver {
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
			sf: sf,
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
		sf: sf,
	}
}

func (d *Driver) Name() string {
	return "redis"
}

func (d *Driver) Initialize() (cache.Cache, error) {
	return newRedisCache(d.cfg, d.sf)
}
