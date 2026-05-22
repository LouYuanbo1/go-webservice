package gormx

type PointerModel[T any] interface {
	*T
}

func IsZero[ID comparable](id ID) bool {
	var zero ID
	return id == zero
}
