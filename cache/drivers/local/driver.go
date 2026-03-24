package local

import "github.com/LouYuanbo1/go-webservice/cache"

/*
这是一个非常经典且重要的 Go 语言技巧：编译时接口实现检查（Compile-time Interface Compliance Check）。
这行代码的作用是：在编译阶段强制验证 *Driver 类型是否完整实现了 cache.Driver 接口。
如果没实现，编译直接报错，而不是等到运行时才发现问题。
// 确保实现了 cache.Driver 接口
var _ cache.Driver = (*Driver)(nil)
*/

type Driver struct {
	cfg *Config
}

// New 返回一个配置好的 Driver
func New(cfg *Config) *Driver {
	if cfg == nil {
		return &Driver{
			cfg: &Config{
				NumCounters: 1e5,
				MaxCost:     1e4,
				BufferItems: 64,
			},
		}
	}
	return &Driver{
		cfg: &Config{
			NumCounters: cfg.NumCounters,
			MaxCost:     cfg.MaxCost,
			BufferItems: cfg.BufferItems,
		},
	}
}

func (d *Driver) Name() string {
	return "local"
}

func (d *Driver) Initialize() (cache.Cache, error) {
	return newLocalCache(d.cfg)
}
