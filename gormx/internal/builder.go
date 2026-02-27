package internal

import (
	"fmt"

	"github.com/LouYuanbo1/go-webservice/gormx/errors"
	"github.com/LouYuanbo1/go-webservice/gormx/options"
	"gorm.io/gorm/clause"
)

func (gx *gormX[T, ID, PT]) clauseOnConflictBuilder(opts ...options.ConflictOption) (*clause.OnConflict, error) {
	conflict := options.NewConflictWithOptions(opts...)

	strategy := conflict.GetStrategy()
	constraintName := conflict.GetConstraintName()
	constraintColumns := conflict.GetConstraintColumns()
	updateColumns := conflict.GetUpdateColumns()

	if len(constraintColumns) == 0 && constraintName == "" {
		return nil, errors.NewWithDetails(
			errors.ErrEmptyConstraint,
			"Validate",
			"",
			"constraint columns or name must be specified",
			nil,
		)
	}

	switch strategy {
	case options.ConflictUpdateColumns:
		if len(updateColumns) == 0 {
			return nil, errors.NewWithDetails(
				errors.ErrEmptyUpdateColumns,
				"Validate",
				"",
				"update columns must be specified when using ConflictUpdateColumns strategy",
				nil,
			)
		}
	}

	clauseConflict := &clause.OnConflict{}

	// 设置约束条件
	if constraintName != "" {
		clauseConflict.OnConstraint = constraintName
	} else if len(constraintColumns) > 0 {
		clauseConflict.Columns = make([]clause.Column, len(constraintColumns))
		for i, col := range constraintColumns {
			clauseConflict.Columns[i] = clause.Column{Name: col}
		}
	}
	// 设置策略
	switch strategy {
	case options.ConflictDoNothing:
		clauseConflict.DoNothing = true
	case options.ConflictUpdateColumns:
		clauseConflict.DoUpdates = clause.AssignmentColumns(updateColumns)
	case options.ConflictUpdateAll:
		clauseConflict.UpdateAll = true
	default:
		return nil, errors.NewWithDetails(
			errors.ErrInvalidConflictStrategy,
			"Build",
			"",
			fmt.Sprintf("unknown conflict strategy: %d", strategy),
			nil,
		)
	}
	return clauseConflict, nil
}

func (gx *gormX[T, ID, PT]) clauseOrderBuilder(opts ...options.OrderOption) *clause.OrderBy {
	order := options.NewOrderWithOptions(opts...)
	columns := order.GetColumns()
	if len(columns) == 0 {
		return nil
	}
	orderBy := &clause.OrderBy{
		Columns: make([]clause.OrderByColumn, 0, len(columns)),
	}
	for _, col := range columns {
		orderBy.Columns = append(orderBy.Columns, clause.OrderByColumn{
			Column: clause.Column{Name: col.GetColumn()},
			Desc:   col.GetDesc(),
		})
	}
	return orderBy
}
