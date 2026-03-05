package imgutil

import (
	"errors"
)

var (
	ErrLoadImage   = errors.New("imgutil: load image error")
	ErrSaveImage   = errors.New("imgutil: save image error")
	ErrDeleteImage = errors.New("imgutil: delete image error")
)
