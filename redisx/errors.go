package redisx

import (
	"errors"
)

var (
	ErrJsonMarshal   = errors.New("redisx: json marshal error")
	ErrJsonUnmarshal = errors.New("redisx: json unmarshal error")
	ErrExpire        = errors.New("redisx: expire error")
	ErrSet           = errors.New("redisx: set error")
	ErrHSet          = errors.New("redisx: hset error")
	ErrGet           = errors.New("redisx: get error")
	ErrHGet          = errors.New("redisx: hget error")
	ErrHMGet         = errors.New("redisx: hmget error")
	ErrHGetAll       = errors.New("redisx: hgetall error")
	ErrNewDecoder    = errors.New("redisx: new decoder error")
	ErrDecode        = errors.New("redisx: decode error")
	ErrDel           = errors.New("redisx: del error")
	ErrAcquire       = errors.New("redisx: acquire error")
	ErrRelease       = errors.New("redisx: release error")
)
