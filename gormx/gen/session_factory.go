package gen

import "gorm.io/gorm"

type SessionFactory struct {
	db *gorm.DB // 仅内部持有，不对外暴露
}

func NewSessionFromFactory[T any, ID comparable, PT PointerModel[T, ID]](sf *SessionFactory) Session[T, ID, PT] {
	return NewSession[T, ID, PT](sf.db)
}
