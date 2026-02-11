package options

type Cost struct {
	cost int
}

func NewCost() *Cost {
	return &Cost{}
}

func (c *Cost) GetCost() int {
	return c.cost
}

func (c *Cost) WithCost(cost int) *Cost {
	c.cost = cost
	return c
}

type CostOption func(*Cost)

func WithCostOption(cost int) CostOption {
	return func(c *Cost) {
		c.cost = cost
	}
}

func NewCostWithOptions(opts ...CostOption) *Cost {
	c := NewCost()
	for _, opt := range opts {
		opt(c)
	}
	return c
}
