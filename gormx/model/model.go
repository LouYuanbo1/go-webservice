package model

/*
GetID() returns the primary key value of the model.
GetPrimaryKey() returns the primary key name of the model.
TableName() returns the table name of the model.

Example:

type User struct {
	ID        uint64    `gorm:"primaryKey" redis:"id"`
	Name      string    `gorm:"not null" redis:"name"`
	Email     string    `gorm:"not null;unique" redis:"email"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`
	UpdatedAt time.Time `gorm:"not null;default:current_timestamp"`
}

func (u *User) GetID() uint64 {
	return u.ID
}

func (u *User) GetPrimaryKey() string {
	return "id"
}

func (u *User) TableName() string {
	return "users"
}
*/
type Model interface {
	GetID() uint64
	GetPrimaryKey() string
	TableName() string
}

/*
PointerModel 定义了一个指针类型的模型接口
它要求T必须是一个指针类型,并嵌入了Model接口
我们使用它来进行约束,确保只有我们希望的指针类型可以实现该interface
例如,我们可以定义一个UserRepository接口,要求其方法参数只能是*User类型

PointerModel defines a pointer type model interface.
It requires T to be a pointer type and embeds the Model interface.
We use it to impose constraints, ensuring that only the pointer types we want can implement the interface.
For example, we can define a UserRepository interface that requires its methods to accept only *User types.
*/
type PointerModel[T any] interface {
	*T
	//Embed existing interfaces
	Model
}
