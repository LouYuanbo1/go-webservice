package internal

import (
	"log"

	"github.com/LouYuanbo1/go-webservice/imgutil/config"
	"github.com/LouYuanbo1/go-webservice/imgutil/options"
	"github.com/disintegration/imaging"
)

type transform struct {
	height int
	width  int
	filter imaging.ResampleFilter
}

func (i *imgUtil) transformBuilder(opts ...options.TransformOption) transform {
	cfg := options.NewTransformByConfig(&i.config.Transform)
	for _, opt := range opts {
		opt(cfg)
	}
	t := transform{
		height: cfg.GetHeight(),
		width:  cfg.GetWidth(),
	}
	switch cfg.GetFilter() {
	case config.Lanczos:
		t.filter = imaging.Lanczos
	case config.CatmullRom:
		t.filter = imaging.CatmullRom
	case config.MitchellNetravali:
		t.filter = imaging.MitchellNetravali
	case config.Linear:
		t.filter = imaging.Linear
	case config.Box:
		t.filter = imaging.Box
	case config.NearestNeighbor:
		t.filter = imaging.NearestNeighbor
	default:
		log.Printf("unknown filter %v, use lanczos instead", cfg.GetFilter())
		t.filter = imaging.Lanczos
	}
	return t
}
