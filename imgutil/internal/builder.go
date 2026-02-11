package internal

import (
	"log"

	"github.com/LouYuanbo1/go-webservice/imgutil/options"
	"github.com/disintegration/imaging"
)

type transform struct {
	height int
	width  int
	filter imaging.ResampleFilter
}

func (i *imgUtil) transformBuilder(opts ...options.TransformOption) transform {
	config := options.NewTransform().WithHeight(i.DefaultHeight).WithWidth(i.DefaultWidth).WithFilter(options.Lanczos)
	for _, opt := range opts {
		opt(config)
	}
	t := transform{
		height: config.GetHeight(),
		width:  config.GetWidth(),
	}
	switch config.GetFilter() {
	case options.Lanczos:
		t.filter = imaging.Lanczos
	case options.CatmullRom:
		t.filter = imaging.CatmullRom
	case options.MitchellNetravali:
		t.filter = imaging.MitchellNetravali
	case options.Linear:
		t.filter = imaging.Linear
	case options.Box:
		t.filter = imaging.Box
	case options.NearestNeighbor:
		t.filter = imaging.NearestNeighbor
	default:
		log.Printf("unknown filter %v, use lanczos instead", config.GetFilter())
		t.filter = imaging.Lanczos
	}
	return t
}

func (i *imgUtil) saveBuilder(opts ...options.SaveOption) *options.Save {
	s := options.NewSave().WithStorageDir(i.DefaultStorageDir).WithQuality(i.DefaultQuality)
	for _, opt := range opts {
		opt(s)
	}
	return s
}
