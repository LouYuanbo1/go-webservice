package options

import (
	"github.com/LouYuanbo1/go-webservice/imgutil/config"
)

type transform struct {
	height int
	width  int
	filter config.Filter
}

func (t *transform) GetHeight() int {
	return t.height
}

func (t *transform) GetWidth() int {
	return t.width
}

func (t *transform) GetFilter() config.Filter {
	return t.filter
}

func NewTransform() *transform {
	return &transform{}
}

func NewTransformWithOptions(opts ...TransformOption) *transform {
	t := NewTransform()
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func NewTransformByConfig(config *config.TransformConfig) *transform {
	return &transform{}
}

//链式调用
func (t *transform) WithHeight(height int) *transform {
	t.height = height
	return t
}

func (t *transform) WithWidth(width int) *transform {
	t.width = width
	return t
}

func (t *transform) WithFilter(filter config.Filter) *transform {
	t.filter = filter
	return t
}

type TransformOption func(*transform)

func WithHeight(height int) TransformOption {
	return func(t *transform) {
		t.height = height
	}
}

func WithWidth(width int) TransformOption {
	return func(t *transform) {
		t.width = width
	}
}

func WithFilter(filter config.Filter) TransformOption {
	return func(t *transform) {
		t.filter = filter
	}
}
