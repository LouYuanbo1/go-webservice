package options

import "gorm.io/gorm/clause"

type Order struct {
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
func NewOrder() *Order {
	return &Order{
		columns: make([]orderColumn, 0),
	}
}

// WithColumn 链式调用方法，添加排序列
func (o *Order) WithColumn(column string, desc bool) *Order {
	o.columns = append(o.columns, orderColumn{column: column, desc: desc})
	return o
}

// WithAsc 链式调用方法，添加升序列
func (o *Order) WithAsc(column string) *Order {
	return o.WithColumn(column, false)
}

// WithDesc 链式调用方法，添加降序列
func (o *Order) WithDesc(column string) *Order {
	return o.WithColumn(column, true)
}

// Build 构建clause.OrderBy
func (o *Order) Build() *clause.OrderBy {
	if len(o.columns) == 0 {
		return nil
	}

	orderBy := &clause.OrderBy{
		Columns: make([]clause.OrderByColumn, 0, len(o.columns)),
	}

	for _, col := range o.columns {
		orderBy.Columns = append(orderBy.Columns, clause.OrderByColumn{
			Column: clause.Column{Name: col.column},
			Desc:   col.desc,
		})
	}

	return orderBy
}

/*
以下是为了支持函数式选项模式而定义的函数类型和函数
e.g.
result, err := gx.FindWithOrder(
    options.WithAscOption("created_at"),
    options.WithDescOption("priority"),
)
*/
type OrderOption func(*Order)

// WithColumnOption 函数式选项，添加排序列
func WithColumnOption(column string, desc bool) OrderOption {
	return func(o *Order) {
		o.WithColumn(column, desc)
	}
}

// WithAscOption 函数式选项，添加升序列
func WithAscOption(column string) OrderOption {
	return WithColumnOption(column, false)
}

// WithDescOption 函数式选项，添加降序列
func WithDescOption(column string) OrderOption {
	return WithColumnOption(column, true)
}

func NewOrderWithOptions(opts ...OrderOption) *Order {
	order := NewOrder()
	for _, opt := range opts {
		opt(order)
	}
	return order
}
