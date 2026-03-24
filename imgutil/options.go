package imgutil

import (
	"log"
	"github.com/disintegration/imaging"
)

type save struct {
	storageDir string
	quality    int
}

func newSaveByConfig(config SaveConfig) *save {
	return &save{
		storageDir: config.StorageDir,
		quality:    config.Quality,
	}
}

type SaveOption func(*save)

func WithStorageDir(dir string) SaveOption {
	return func(s *save) {
		s.storageDir = dir
	}
}

func WithQuality(quality int) SaveOption {
	return func(s *save) {
		s.quality = quality
	}
}

func (i *imgUtil) saveBuilder(opts ...SaveOption) *save {
	s := newSaveByConfig(i.config.Save)
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type transform struct {
	height int
	width  int
	filter Filter
}

func newTransformByConfig(config TransformConfig) *transform {
	return &transform{
		height: config.Height,
		width:  config.Width,
		filter: config.Filter,
	}
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

func WithFilter(filter Filter) TransformOption {
	return func(t *transform) {
		t.filter = filter
	}
}

type transformWithFilter struct {
	height int
	width  int
	filter imaging.ResampleFilter
}

func (i *imgUtil) transformBuilder(opts ...TransformOption) transformWithFilter {
	transform := newTransformByConfig(i.config.Transform)
	for _, opt := range opts {
		opt(transform)
	}
	t := transformWithFilter{
		height: transform.height,
		width:  transform.width,
	}
	switch transform.filter {
	case Lanczos:
		t.filter = imaging.Lanczos
	case CatmullRom:
		t.filter = imaging.CatmullRom
	case MitchellNetravali:
		t.filter = imaging.MitchellNetravali
	case Linear:
		t.filter = imaging.Linear
	case Box:
		t.filter = imaging.Box
	case NearestNeighbor:
		t.filter = imaging.NearestNeighbor
	default:
		log.Printf("unknown filter %v, use lanczos instead", transform.filter)
		t.filter = imaging.Lanczos
	}
	return t
}
