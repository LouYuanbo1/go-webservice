package cache

type Driver interface {
	// Name 返回驱动名称，用于日志或标识
	Name() string
	// Initialize 初始化并返回具体的 Cache 实现
	Initialize() (Cache, error)
}
