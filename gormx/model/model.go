package model

type Model interface {
	GetID() uint64
	GetPrimaryKey() string
	TableName() string
}

// PointerModel 定义了一个指针类型的模型接口
// PointerModel defines a pointer type model interface.
// 它要求T必须是一个指针类型,并嵌入了Model接口
// It requires T to be a pointer type and embeds the Model interface.
// 我们使用它来进行约束,确保只有我们希望的指针类型可以实现该interface
// We use it to impose constraints, ensuring that only the pointer types we want can implement the interface.
// 例如,我们可以定义一个UserRepository接口,要求其方法参数只能是*User类型
// For example, we can define a UserRepository interface that requires its methods to accept only *User types.
type PointerModel[T any] interface {
	*T    // 核心：T必须是一个指针类型
	Model // 嵌入原有接口
}
