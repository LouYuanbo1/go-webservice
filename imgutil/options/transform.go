package options

type Filter int

const (
	Lanczos Filter = iota
	CatmullRom
	MitchellNetravali
	Linear
	Box
	NearestNeighbor
)

type Transform struct {
	height int
	width  int
	filter Filter
}

func NewTransform() *Transform {
	return &Transform{}
}

func (t *Transform) GetHeight() int {
	return t.height
}

func (t *Transform) GetWidth() int {
	return t.width
}

func (t *Transform) GetFilter() Filter {
	return t.filter
}

//链式调用

func (t *Transform) WithHeight(height int) *Transform {
	t.height = height
	return t
}

func (t *Transform) WithWidth(width int) *Transform {
	t.width = width
	return t
}

func (t *Transform) WithFilter(filter Filter) *Transform {
	t.filter = filter
	return t
}

type TransformOption func(*Transform)

func WithHeight(height int) TransformOption {
	return func(t *Transform) {
		t.height = height
	}
}

func WithWidth(width int) TransformOption {
	return func(t *Transform) {
		t.width = width
	}
}

func WithFilter(filter Filter) TransformOption {
	return func(t *Transform) {
		t.filter = filter
	}
}

func NewTransformWithOptions(opts ...TransformOption) *Transform {
	t := NewTransform()
	for _, opt := range opts {
		opt(t)
	}
	return t
}
