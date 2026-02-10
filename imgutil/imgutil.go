package imgutil

import (
	"image"

	"github.com/LouYuanbo1/go-webservice/imgutil/config"
	"github.com/LouYuanbo1/go-webservice/imgutil/internal"
	"github.com/LouYuanbo1/go-webservice/imgutil/options"
)

type ImgUtil interface {
	Thumbnail(img image.Image, opts ...options.TransformOption) image.Image
	Save(img image.Image, filename string, opts ...options.SaveOption) error
}

func NewImgUtil(config config.ImgUtilConfig) ImgUtil {
	return internal.NewImgUtil(config)
}
