package imgutil

import (
	"image"

	"github.com/LouYuanbo1/go-webservice/imgutil/config"
	"github.com/LouYuanbo1/go-webservice/imgutil/internal"
	"github.com/LouYuanbo1/go-webservice/imgutil/options"
)

type ImgUtil interface {
	Load(imgPath string) (image.Image, error)
	Thumbnail(img image.Image, opts ...options.TransformOption) image.Image
	Save(img image.Image, filename string, opts ...options.SaveOption) error
	Delete(imgPath string) error
	WithTimestamp(imgPath string, format string) string
}

func NewImgUtil(config config.ImgUtilConfig) ImgUtil {
	return internal.NewImgUtil(config)
}
