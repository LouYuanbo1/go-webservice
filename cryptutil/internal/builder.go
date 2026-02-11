package internal

import "github.com/LouYuanbo1/go-webservice/cryptutil/options"

func (c *cryptUtil) costBuilder(opts ...options.CostOption) *options.Cost {
	cost := options.NewCost().WithCost(c.defaultCost)
	for _, opt := range opts {
		opt(cost)
	}
	return cost
}
