package internal

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LouYuanbo1/go-webservice/imgutil/config"
	"github.com/LouYuanbo1/go-webservice/imgutil/options"
	"github.com/disintegration/imaging"
)

type imgUtil struct {
	DefaultWidth      int    // 默认处理宽度
	DefaultHeight     int    // 默认处理高度
	DefaultQuality    int    // JPEG质量 (1-100)
	DefaultStorageDir string // 存储目录
}

func NewImgUtil(imgUtilConfig config.ImgUtilConfig) *imgUtil {
	return &imgUtil{
		DefaultWidth:      imgUtilConfig.DefaultWidth,
		DefaultHeight:     imgUtilConfig.DefaultHeight,
		DefaultQuality:    imgUtilConfig.DefaultQuality,
		DefaultStorageDir: imgUtilConfig.DefaultStorageDir,
	}
}

// 加载图片
func (i *imgUtil) Load(imgPath string) (image.Image, error) {
	img, err := imaging.Open(imgPath)
	if err != nil {
		return nil, fmt.Errorf("load image %s failed: %w", imgPath, err)
	}
	return img, nil
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
func (i *imgUtil) Save(img image.Image, filename string, opts ...options.SaveOption) error {
	save := i.saveBuilder(opts...)
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

func (i *imgUtil) Delete(imgPath string) error {
	err := os.Remove(imgPath)
	if err != nil {
		return fmt.Errorf("delete image %s failed: %w", imgPath, err)
	}
	return nil
}

func (i *imgUtil) WithFormatTimestamp(imgPath string, format string) string {
	//获取时间戳
	timestamp := time.Now().Format(format)
	//获取基础文件名,去掉上层文件夹
	filename := filepath.Base(imgPath)
	//获取文件类型
	ext := filepath.Ext(filename)
	//去掉文件名的扩展名
	filename = strings.TrimSuffix(filename, ext)
	//添加时间戳
	return filepath.Join(filepath.Dir(imgPath), fmt.Sprintf("%s_%s%s", filename, timestamp, ext))
}

func (i *imgUtil) WithUnixNanoTimestamp(imgPath string) string {
	//获取时间戳
	timestamp := time.Now().UnixNano()
	//获取基础文件名,去掉上层文件夹
	filename := filepath.Base(imgPath)
	//获取文件类型
	ext := filepath.Ext(filename)
	//去掉文件名的扩展名
	filename = strings.TrimSuffix(filename, ext)
	//添加时间戳
	return filepath.Join(filepath.Dir(imgPath), fmt.Sprintf("%s_%d%s", filename, timestamp, ext))
}
