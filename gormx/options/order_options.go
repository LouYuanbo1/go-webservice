package options

type order struct {
	columns []orderColumn
}

type orderColumn struct {
	column string
	desc   bool
}

func (o *order) GetColumns() []orderColumn {
	return o.columns
}

func (oc *orderColumn) GetColumn() string {
	return oc.column
}

func (oc *orderColumn) GetDesc() bool {
	return oc.desc
}

/*
链式调用
order := options.NewOrder().WithAsc("created_at").WithDesc("priority")
然后将order传递给需要的方法，或者直接使用order.Build()
*/

// NewOrder 创建一个新的Order实例
func NewOrder() *order {
	return &order{
		columns: make([]orderColumn, 0),
	}
}

func NewOrderWithOptions(opts ...OrderOption) *order {
	order := NewOrder()
	for _, opt := range opts {
		opt(order)
	}
	return order
}

// WithColumn 链式调用方法，添加排序列
func (o *order) WithColumn(column string, desc bool) *order {
	o.columns = append(o.columns, orderColumn{column: column, desc: desc})
	return o
}

// WithAsc 链式调用方法，添加升序列
func (o *order) WithAsc(column string) *order {
	return o.WithColumn(column, false)
}

// WithDesc 链式调用方法，添加降序列
func (o *order) WithDesc(column string) *order {
	return o.WithColumn(column, true)
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
		o.WithColumn(column, desc)
	}
}

// WithAscOption 函数式选项，添加升序列
func WithAsc(column string) OrderOption {
	return WithColumn(column, false)
}

// WithDescOption 函数式选项，添加降序列
func WithDesc(column string) OrderOption {
	return WithColumn(column, true)
}
