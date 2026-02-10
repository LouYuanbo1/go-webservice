package internal

import (
	"image"
	"image/png"
	"log"
	"path/filepath"
	"strings"

	"github.com/LouYuanbo1/go-webservice/imgutil/config"
	"github.com/LouYuanbo1/go-webservice/imgutil/options"
	"github.com/disintegration/imaging"
)

type imgUtil struct {
	/*
		MaxFileSize      int64    // 最大文件大小（字节）
		AllowedMimeTypes []string // 允许的MIME类型
	*/
	DefaultWidth      int    // 默认处理宽度
	DefaultHeight     int    // 默认处理高度
	DefaultQuality    int    // JPEG质量 (1-100)
	DefaultStorageDir string // 存储目录
}

func NewImgUtil(imgUtilConfig config.ImgUtilConfig) *imgUtil {
	return &imgUtil{
		/*
			MaxFileSize:      imgUtilConfig.MaxFileSize,
			AllowedMimeTypes: imgUtilConfig.AllowedMimeTypes,
		*/
		DefaultWidth:      imgUtilConfig.DefaultWidth,
		DefaultHeight:     imgUtilConfig.DefaultHeight,
		DefaultQuality:    imgUtilConfig.DefaultQuality,
		DefaultStorageDir: imgUtilConfig.DefaultStorageDir,
	}
}

type transform struct {
	height int
	width  int
	filter imaging.ResampleFilter
}

func (i *imgUtil) transformBuilder(opts ...options.TransformOption) transform {
	config := options.Transform{
		Height: i.DefaultHeight,
		Width:  i.DefaultWidth,
		Filter: options.Lanczos,
	}
	for _, opt := range opts {
		opt(&config)
	}
	t := transform{
		height: config.Height,
		width:  config.Width,
	}
	switch config.Filter {
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
		log.Printf("unknown filter %v, use lanczos instead", config.Filter)
		t.filter = imaging.Lanczos
	}
	return t
}

func (i *imgUtil) saveBuilder(opts ...options.SaveOption) options.Save {
	s := options.Save{
		StorageDir: i.DefaultStorageDir,
		Quality:    i.DefaultQuality,
	}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}

func (i *imgUtil) Thumbnail(img image.Image, opts ...options.TransformOption) image.Image {
	t := i.transformBuilder(opts...)
	img = imaging.Thumbnail(img, t.width, t.height, t.filter)
	return img
}

// 保存图片,按照配置的质量保存
func (s *imgUtil) Save(img image.Image, filename string, opts ...options.SaveOption) error {
	save := s.saveBuilder(opts...)
	ext := strings.ToLower(filepath.Ext(filename))
	fullPath := filepath.Join(save.StorageDir, filename)
	switch ext {
	case ".jpg", ".jpeg":
		return imaging.Save(img, fullPath, imaging.JPEGQuality(save.Quality))
	case ".png":
		level := save.Quality * 9 / 100
		level = max(level, 1)
		level = min(level, 9)
		return imaging.Save(img, fullPath, imaging.PNGCompressionLevel(png.CompressionLevel(level)))
	default:
		return imaging.Save(img, fullPath)
	}
}
