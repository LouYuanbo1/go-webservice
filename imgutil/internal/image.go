package internal

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LouYuanbo1/go-webservice/imgutil/config"
	"github.com/LouYuanbo1/go-webservice/imgutil/errors"
	"github.com/LouYuanbo1/go-webservice/imgutil/options"
	"github.com/disintegration/imaging"
)

type imgUtil struct {
	config config.ImgUtilConfig
}

func NewImgUtil(imgUtilConfig config.ImgUtilConfig) *imgUtil {
	return &imgUtil{
		config: imgUtilConfig,
	}
}

// 加载图片
func (i *imgUtil) Load(imgPath string) (image.Image, error) {
	img, err := imaging.Open(imgPath)
	if err != nil {
		return nil, errors.New(
			errors.ErrLoadImage,
			"Load",
			err,
		)
	}
	return img, nil
}

func (i *imgUtil) Thumbnail(img image.Image, opts ...options.TransformOption) image.Image {
	t := i.transformBuilder(opts...)
	img = imaging.Thumbnail(img, t.width, t.height, t.filter)
	return img
}

// 保存图片,按照配置的质量保存
func (i *imgUtil) Save(img image.Image, filename string, opts ...options.SaveOption) error {
	save := options.SaveBuilder(&i.config.Save, opts...)

	ext := strings.ToLower(filepath.Ext(filename))
	fullPath := filepath.Join(save.GetStorageDir(), filename)
	var err error
	switch ext {
	case ".jpg", ".jpeg":
		err = imaging.Save(img, fullPath, imaging.JPEGQuality(save.GetQuality()))
	case ".png":
		level := save.GetQuality() * 9 / 100
		level = max(level, 1)
		level = min(level, 9)
		err = imaging.Save(img, fullPath, imaging.PNGCompressionLevel(png.CompressionLevel(level)))
	default:
		err = imaging.Save(img, fullPath)
	}
	if err != nil {
		return errors.New(
			errors.ErrSaveImage,
			"Save",
			err,
		)
	}
	return nil
}

func (i *imgUtil) Delete(imgPath string) error {
	err := os.Remove(imgPath)
	if err != nil {
		return errors.New(
			errors.ErrDeleteImage,
			"Delete",
			err,
		)
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
