package options

type Cost struct {
	Value int
}

type CostOption func(*Cost)

func WithCost(value int) CostOption {
	return func(c *Cost) {
		c.Value = value
	}
}
