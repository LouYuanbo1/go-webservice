package options

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

// GetStrategy 获取冲突处理策略
func (c *conflict) GetStrategy() ConflictStrategy {
	return c.strategy
}

func (c *conflict) GetConstraintName() string {
	return c.constraintName
}

func (c *conflict) GetConstraintColumns() []string {
	return c.constraintColumns
}

func (c *conflict) GetUpdateColumns() []string {
	return c.updateColumns
}

// NewConflict 创建新的冲突配置
func NewConflict() *conflict {
	return &conflict{}
}

// NewConflictWithOptions 使用函数式选项创建冲突配置
func NewConflictWithOptions(opts ...ConflictOption) *conflict {
	c := NewConflict()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// 链式调用方法
// DoNothing 设置冲突时不执行任何操作
func (c *conflict) DoNothing() *conflict {
	c.strategy = ConflictDoNothing
	return c
}

// UpdateAll 设置冲突时更新所有列
func (c *conflict) UpdateAll() *conflict {
	c.strategy = ConflictUpdateAll
	return c
}

// WithConstraintName 设置约束名
func (c *conflict) WithConstraintName(name string) *conflict {
	c.constraintName = name
	return c
}

// WithConstraintColumns 设置约束列
func (c *conflict) WithConstraintColumns(columns ...string) *conflict {
	c.constraintColumns = columns
	return c
}

// UpdateColumns 设置冲突时更新指定列
func (c *conflict) UpdateColumns(columns ...string) *conflict {
	c.strategy = ConflictUpdateColumns
	c.updateColumns = columns
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

// UpdateColumnsOption 函数式选项 - 设置冲突时更新指定列
func UpdateColumns(columns ...string) ConflictOption {
	return func(c *conflict) {
		c.strategy = ConflictUpdateColumns
		c.updateColumns = columns
	}
}
