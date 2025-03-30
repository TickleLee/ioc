package scope

// Scope 定义实例的生命周期类型
type Scope int

const (
	// Singleton 表示全局单例，容器中只有一个实例
	Singleton Scope = iota

	// Prototype 表示多例模式，每次获取依赖时创建新实例
	Prototype
)
