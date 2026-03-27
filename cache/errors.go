package cache

import (
	"errors"
)

var (
	WarnSetCacheFailed = errors.New("cache: set cache failed")
)

var (
	ErrDriverNotFound = errors.New("cache: driver not found")
	ErrInit          = errors.New("cache: init error")
	ErrJsonMarshal   = errors.New("cache: json marshal error")
	ErrJsonUnmarshal = errors.New("cache: json unmarshal error")
	ErrExpire        = errors.New("cache: expire error")
	ErrSet           = errors.New("cache: set error")
	ErrGet           = errors.New("cache: get error")
	ErrDel           = errors.New("cache: del error")
)
