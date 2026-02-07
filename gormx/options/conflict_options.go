package options

const (
	ConflictDoNothing ConflictStrategy = iota
	ConflictUpdateColumns
	ConflictUpdateAll
)

type ConflictStrategy int
type ConflictOption func(*Conflict)

// OnConflictOption 冲突处理选项
type Conflict struct {
	Strategy          ConflictStrategy
	ConstraintName    string
	ConstraintColumns []string
	UpdateColumns     []string
}

func DoNothing() ConflictOption {
	return func(o *Conflict) {
		o.Strategy = ConflictDoNothing
	}
}

func UpdateColumns(columns ...string) ConflictOption {
	return func(o *Conflict) {
		o.Strategy = ConflictUpdateColumns
		o.UpdateColumns = columns
	}
}

func UpdateAll() ConflictOption {
	return func(o *Conflict) {
		o.Strategy = ConflictUpdateAll
	}
}

func Constraint(name string) ConflictOption {
	return func(o *Conflict) {
		o.ConstraintName = name
	}
}

func ConstraintColumns(columns ...string) ConflictOption {
	return func(o *Conflict) {
		o.ConstraintColumns = columns
	}
}
