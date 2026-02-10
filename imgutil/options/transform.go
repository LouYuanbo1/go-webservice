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
	Height int
	Width  int
	Filter Filter
}

type TransformOption func(*Transform)

func WithHeight(height int) TransformOption {
	return func(t *Transform) {
		t.Height = height
	}
}

func WithWidth(width int) TransformOption {
	return func(t *Transform) {
		t.Width = width
	}
}

func WithFilter(filter Filter) TransformOption {
	return func(t *Transform) {
		t.Filter = filter
	}
}
