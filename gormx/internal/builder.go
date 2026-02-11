package internal

import (
	"github.com/LouYuanbo1/go-webservice/gormx/options"
	"gorm.io/gorm/clause"
)

func (gx *gormX[T, ID, PT]) clauseOnConflictBuilder(opts ...options.ConflictOption) (*clause.OnConflict, error) {
	conflict := options.NewConflictWithOptions(opts...)
	return conflict.Build()
}

func (gx *gormX[T, ID, PT]) clauseOrderBuilder(opts ...options.OrderOption) *clause.OrderBy {
	order := options.NewOrderWithOptions(opts...)
	return order.Build()
}