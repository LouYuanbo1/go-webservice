package options

import "github.com/LouYuanbo1/go-webservice/cryptutil/config"

type cost struct {
	value int
}

func (c *cost) GetValue() int {
	return c.value
}

func NewCost() *cost {
	return &cost{}
}

func NewCostWithOptions(opts ...CostOption) *cost {
	c := NewCost()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func NewCostByConfig(cfg config.CryptUtilConfig) *cost {
	return &cost{
		value: cfg.Cost,
	}
}

func CostBuilder(cfg config.CryptUtilConfig, opts ...CostOption) *cost {
	c := NewCostByConfig(cfg)
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *cost) WithCost(value int) {
	c.value = value
}

type CostOption func(*cost)

func WithCost(value int) CostOption {
	return func(c *cost) {
		c.value = value
	}
}
