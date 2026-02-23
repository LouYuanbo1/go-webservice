package model

import "github.com/elastic/go-elasticsearch/v9/typedapi/types"

type Document interface {
	Index() string
	GetStringID() string
	GetTypeMapping() *types.TypeMapping
}

type PointerDocument[T any] interface {
	*T
	Document
}
