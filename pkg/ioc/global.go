package ioc

import (
	"sync"
)

var (
	// 全局默认容器实例
	defaultContainer Container

	// 保证全局容器只初始化一次的锁
	once sync.Once
)

// 获取默认的容器实例
func getDefaultContainer() Container {
	once.Do(func() {
		defaultContainer = NewContainer()
	})
	return defaultContainer
}

// Register 注册依赖到默认容器
func Register(name string, instance interface{}, scope Scope) error {
	return getDefaultContainer().Register(name, instance, scope)
}

// RegisterType 按类型注册依赖到默认容器
func RegisterType(typeName string, instance interface{}) error {
	return getDefaultContainer().RegisterType(typeName, instance)
}

// RegisterFactory 注册依赖工厂到默认容器
func RegisterFactory(name string, scope Scope, factory func() (interface{}, error)) error {
	return getDefaultContainer().RegisterFactory(name, scope, factory)
}

// Get 从默认容器获取依赖
func Get(name string) interface{} {
	return getDefaultContainer().Get(name)
}

// GetSafe 安全地从默认容器获取依赖
func GetSafe(name string) (interface{}, error) {
	return getDefaultContainer().GetSafe(name)
}

// GetByType 按类型从默认容器获取依赖
func GetByType(typeName string, name string) interface{} {
	return getDefaultContainer().GetByType(typeName, name)
}

// Inject 注入依赖到指定实例
func Inject(instance interface{}) error {
	return getDefaultContainer().Inject(instance)
}

// Init 初始化默认容器
func Init() error {
	return getDefaultContainer().Init()
}
