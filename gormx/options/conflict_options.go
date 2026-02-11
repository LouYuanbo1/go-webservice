package options

import (
	"fmt"

	"github.com/LouYuanbo1/go-webservice/gormx/errors"
	"gorm.io/gorm/clause"
)

const (
	ConflictDoNothing ConflictStrategy = iota
	ConflictUpdateColumns
	ConflictUpdateAll
)

type ConflictStrategy int

// Conflict 冲突处理配置
type Conflict struct {
	strategy          ConflictStrategy
	constraintName    string
	constraintColumns []string
	updateColumns     []string
}

// NewConflict 创建新的冲突配置
func NewConflict() *Conflict {
	return &Conflict{}
}

// 链式调用方法
// WithConstraintName 设置约束名
func (c *Conflict) WithConstraintName(name string) *Conflict {
	c.constraintName = name
	return c
}

// WithConstraintColumns 设置约束列
func (c *Conflict) WithConstraintColumns(columns ...string) *Conflict {
	c.constraintColumns = columns
	return c
}

// DoNothing 设置冲突时不执行任何操作
func (c *Conflict) DoNothing() *Conflict {
	c.strategy = ConflictDoNothing
	return c
}

// UpdateColumns 设置冲突时更新指定列
func (c *Conflict) UpdateColumns(columns ...string) *Conflict {
	c.strategy = ConflictUpdateColumns
	c.updateColumns = columns
	return c
}

// UpdateAll 设置冲突时更新所有列
func (c *Conflict) UpdateAll() *Conflict {
	c.strategy = ConflictUpdateAll
	return c
}

// Validate 验证配置
func (c *Conflict) Validate() error {
	if len(c.constraintColumns) == 0 && c.constraintName == "" {
		return errors.NewWithDetails(
			errors.ErrEmptyConstraint,
			"Validate",
			"",
			"constraint columns or name must be specified",
			nil,
		)
	}

	switch c.strategy {
	case ConflictUpdateColumns:
		if len(c.updateColumns) == 0 {
			return errors.NewWithDetails(
				errors.ErrEmptyUpdateColumns,
				"Validate",
				"",
				"update columns must be specified when using ConflictUpdateColumns strategy",
				nil,
			)
		}
	}

	return nil
}

// Build 构建GORM OnConflict子句
func (c *Conflict) Build() (*clause.OnConflict, error) {
	if err := c.Validate(); err != nil {
		return nil, errors.New(
			errors.ErrInvalidOnConflictClause,
			"Build",
			"",
			err,
		)
	}

	clauseConflict := &clause.OnConflict{}

	// 设置约束条件
	if c.constraintName != "" {
		clauseConflict.OnConstraint = c.constraintName
	} else if len(c.constraintColumns) > 0 {
		clauseConflict.Columns = make([]clause.Column, len(c.constraintColumns))
		for i, col := range c.constraintColumns {
			clauseConflict.Columns[i] = clause.Column{Name: col}
		}
	}

	// 设置策略
	switch c.strategy {
	case ConflictDoNothing:
		clauseConflict.DoNothing = true
	case ConflictUpdateColumns:
		clauseConflict.DoUpdates = clause.AssignmentColumns(c.updateColumns)
	case ConflictUpdateAll:
		clauseConflict.UpdateAll = true
	default:
		return nil, errors.NewWithDetails(
			errors.ErrInvalidConflictStrategy,
			"Build",
			"",
			fmt.Sprintf("unknown conflict strategy: %d", c.strategy),
			nil,
		)
	}

	return clauseConflict, nil
}

// 函数式选项模式
type ConflictOption func(*Conflict)

// OnConstraint 函数式选项 - 设置约束名
func OnConstraint(name string) ConflictOption {
	return func(c *Conflict) {
		c.constraintName = name
	}
}

// OnConstraintColumns 函数式选项 - 设置约束列
func OnConstraintColumns(columns ...string) ConflictOption {
	return func(c *Conflict) {
		c.constraintColumns = columns
	}
}

// DoNothingOption 函数式选项 - 设置冲突时不执行任何操作
func DoNothingOption() ConflictOption {
	return func(c *Conflict) {
		c.strategy = ConflictDoNothing
	}
}

// UpdateColumnsOption 函数式选项 - 设置冲突时更新指定列
func UpdateColumnsOption(columns ...string) ConflictOption {
	return func(c *Conflict) {
		c.strategy = ConflictUpdateColumns
		c.updateColumns = columns
	}
}

// UpdateAllOption 函数式选项 - 设置冲突时更新所有列
func UpdateAllOption() ConflictOption {
	return func(c *Conflict) {
		c.strategy = ConflictUpdateAll
	}
}

// NewConflictWithOptions 使用函数式选项创建冲突配置
func NewConflictWithOptions(opts ...ConflictOption) *Conflict {
	c := NewConflict()
	for _, opt := range opts {
		opt(c)
	}
	return c
}
