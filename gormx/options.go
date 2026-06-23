package gormx

import (
	"fmt"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"gorm.io/gorm/clause"
)

const (
	ConflictDoNothing ConflictStrategy = iota
	ConflictUpdateColumns
	ConflictUpdateAll
)

type ConflictStrategy int

// conflict 冲突处理配置
type conflict struct {
	strategy          ConflictStrategy
	constraintName    string
	constraintColumns []string
	updateColumns     []string
}

// NewConflict 创建新的冲突配置
func newConflict() *conflict {
	return &conflict{}
}

// NewConflictWithOptions 使用函数式选项创建冲突配置
func newConflictWithOptions(opts ...ConflictOption) *conflict {
	c := newConflict()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// 函数式选项模式
type ConflictOption func(*conflict)

// DoNothingOption 函数式选项 - 设置冲突时不执行任何操作
func DoNothing() ConflictOption {
	return func(c *conflict) {
		c.strategy = ConflictDoNothing
	}
}

// UpdateColumnsOption 函数式选项 - 设置冲突时更新指定列
func UpdateColumns(columns ...string) ConflictOption {
	return func(c *conflict) {
		c.strategy = ConflictUpdateColumns
		c.updateColumns = columns
	}
}

// UpdateAllOption 函数式选项 - 设置冲突时更新所有列
func UpdateAll() ConflictOption {
	return func(c *conflict) {
		c.strategy = ConflictUpdateAll
	}
}

// OnConstraint 函数式选项 - 设置约束名
func OnConstraint(name string) ConflictOption {
	return func(c *conflict) {
		c.constraintName = name
	}
}

// OnConstraintColumns 函数式选项 - 设置约束列
func OnConstraintColumns(columns ...string) ConflictOption {
	return func(c *conflict) {
		c.constraintColumns = columns
	}
}

func clauseOnConflictBuilder(opts ...ConflictOption) (*clause.OnConflict, error) {
	conflict := newConflictWithOptions(opts...)

	if len(conflict.constraintColumns) == 0 && conflict.constraintName == "" {
		return nil, errorx.NewWithDetails(
			ErrEmptyConstraint,
			"gormx",
			"clauseOnConflictBuilder",
			"constraint columns or name must be specified",
			nil,
		)
	}

	switch conflict.strategy {
	case ConflictUpdateColumns:
		if len(conflict.updateColumns) == 0 {
			return nil, errorx.NewWithDetails(
				ErrEmptyUpdateColumns,
				"gormx",
				"clauseOnConflictBuilder",
				"update columns must be specified when using ConflictUpdateColumns strategy",
				nil,
			)
		}
	}

	clauseConflict := &clause.OnConflict{}

	// 设置约束条件
	if conflict.constraintName != "" {
		clauseConflict.OnConstraint = conflict.constraintName
	} else if len(conflict.constraintColumns) > 0 {
		clauseConflict.Columns = make([]clause.Column, len(conflict.constraintColumns))
		for i, col := range conflict.constraintColumns {
			clauseConflict.Columns[i] = clause.Column{Name: col}
		}
	}
	// 设置策略
	switch conflict.strategy {
	case ConflictDoNothing:
		clauseConflict.DoNothing = true
	case ConflictUpdateColumns:
		clauseConflict.DoUpdates = clause.AssignmentColumns(conflict.updateColumns)
	case ConflictUpdateAll:
		clauseConflict.UpdateAll = true
	default:
		return nil, errorx.NewWithDetails(
			ErrInvalidConflictStrategy,
			"gormx",
			"clauseOnConflictBuilder",
			fmt.Sprintf("unknown conflict strategy: %d", conflict.strategy),
			nil,
		)
	}
	return clauseConflict, nil
}

type order struct {
	columns []orderColumn
}

type orderColumn struct {
	column string
	desc   bool
}

/*
链式调用
order := options.NewOrder().WithAsc("created_at").WithDesc("priority")
然后将order传递给需要的方法，或者直接使用order.Build()
*/

// NewOrder 创建一个新的Order实例
func newOrder() *order {
	return &order{
		columns: make([]orderColumn, 0),
	}
}

func newOrderWithOptions(opts ...OrderOption) *order {
	order := newOrder()
	for _, opt := range opts {
		opt(order)
	}
	return order
}

/*
以下是为了支持函数式选项模式而定义的函数类型和函数
e.g.
result, err := gx.FindWithOrder(

	options.WithAsc("created_at"),
	options.WithDesc("priority"),

)
*/
type OrderOption func(*order)

// WithColumnOption 函数式选项，添加排序列
func WithColumn(column string, desc bool) OrderOption {
	return func(o *order) {
		o.columns = append(o.columns, orderColumn{column: column, desc: desc})
	}
}

// WithAscOption 函数式选项，添加升序列
func WithAsc(column string) OrderOption {
	return func(o *order) {
		o.columns = append(o.columns, orderColumn{column: column, desc: false})
	}
}

// WithDescOption 函数式选项，添加降序列
func WithDesc(column string) OrderOption {
	return func(o *order) {
		o.columns = append(o.columns, orderColumn{column: column, desc: true})
	}
}

func clauseOrderBuilder(opts ...OrderOption) *clause.OrderBy {
	order := newOrderWithOptions(opts...)
	if len(order.columns) == 0 {
		return nil
	}
	orderBy := &clause.OrderBy{
		Columns: make([]clause.OrderByColumn, 0, len(order.columns)),
	}
	for _, col := range order.columns {
		orderBy.Columns = append(orderBy.Columns, clause.OrderByColumn{
			Column: clause.Column{Name: col.column},
			Desc:   col.desc,
		})
	}
	return orderBy
}
