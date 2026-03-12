package cache

import (
	"errors"
)

var (
	ErrJsonMarshal   = errors.New("cache: json marshal error")
	ErrJsonUnmarshal = errors.New("cache: json unmarshal error")
	ErrExpire        = errors.New("cache: expire error")
	ErrSet           = errors.New("cache: set error")
	ErrHSet          = errors.New("cache: hset error")
	ErrGet           = errors.New("cache: get error")
	ErrHGet          = errors.New("cache: hget error")
	ErrHMGet         = errors.New("cache: hmget error")
	ErrHGetAll       = errors.New("cache: hgetall error")
	ErrNewDecoder    = errors.New("cache: new decoder error")
	ErrDecode        = errors.New("cache: decode error")
	ErrDel           = errors.New("cache: del error")
	ErrAcquire       = errors.New("cache: acquire error")
	ErrRelease       = errors.New("cache: release error")
)
