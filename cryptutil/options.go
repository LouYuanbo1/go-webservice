package cryptutil

type cost struct {
	value int
}

func newCostByConfig(cfg Config) *cost {
	return &cost{
		value: cfg.Cost,
	}
}

type CostOption func(*cost)

func WithCost(value int) CostOption {
	return func(c *cost) {
		c.value = value
	}
}

func (c *cryptUtil) costBuilder(opts ...CostOption) *cost {
	cost := newCostByConfig(c.config)
	for _, opt := range opts {
		opt(cost)
	}
	return cost
}
