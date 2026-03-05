package cryptutil

import (
	"errors"
)

var (
	ErrEncrypt     = errors.New("cryptutil: encrypt error")
	ErrCheckSecret = errors.New("cryptutil: check secret error")
)
