package ioc

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// 引入本地types包的Scope类型
type Scope = int

const (
	// Singleton 表示全局单例，容器中只有一个实例
	Singleton Scope = iota

	// Prototype 表示多例模式，每次获取依赖时创建新实例
	Prototype
)

// 添加初始化阶段常量
const (
	// NotInitialized 表示容器尚未初始化
	NotInitialized = iota
	// InjectionPhase 表示正在进行依赖注入阶段
	InjectionPhase
	// PostConstructPhase 表示正在执行PostConstruct方法阶段
	PostConstructPhase
	// Initialized 表示容器已完全初始化
	Initialized
)

// BeanDefinition 表示容器中注册的对象定义
type BeanDefinition struct {
	// 对象名称
	Name string

	// 类型名称（如 service, repository 等）
	TypeName string

	// 对象的类型
	Type reflect.Type

	// 对象实例（如果是单例）
	Instance interface{}

	// 对象的作用域
	Scope Scope

	// 是否为接口
	IsInterface bool

	// 工厂函数，用于创建对象实例
	Factory func() (interface{}, error)

	// 是否已完成依赖注入
	injected bool

	// 是否已执行PostConstruct
	initialized bool
}

// InitializingBean 接口定义了对象初始化的方法
type InitializingBean interface {
	// PostConstruct 方法在对象被容器创建并注入依赖后调用
	PostConstruct() error
}

// Container 定义IoC容器的接口
type Container interface {
	// 注册依赖到容器
	Register(name string, instance interface{}, scope Scope) error

	// 按类型注册依赖
	RegisterType(typeName string, instance interface{}) error

	// 按类型注册依赖,支持配置名字
	RegisterTypeWithName(typeName string, name string, instance interface{}) error

	// 注册依赖工厂
	RegisterFactory(name string, scope Scope, factory func() (interface{}, error)) error

	// 获取依赖
	Get(name string) interface{}

	// 按类型获取依赖
	GetByType(typeName string, name string) interface{}

	// 安全地获取依赖，返回错误而不是panic
	GetSafe(name string) (interface{}, error)

	// 获取所有注册的bean
	GetAll() map[string]*BeanDefinition

	// 获取所有注册的bean名称
	GetAllNames() []string

	// 注入依赖
	Inject(instance interface{}) error

	// 初始化容器
	Init() error
}

// 默认的容器实现
type containerImpl struct {
	// 所有注册的bean定义
	beans map[string]*BeanDefinition

	// 按类型分组的bean
	typeRegistry map[string]map[string]*BeanDefinition

	// 保护并发访问的互斥锁
	mu sync.RWMutex

	// 标记容器是否已初始化
	initialized bool

	// 用于检测初始化过程中的循环依赖
	initializing map[string]bool

	// 容器当前所处的初始化阶段
	currentPhase int

	// 日志记录器
	logger Logger
}

// 创建新的容器实例
func NewContainer() Container {
	// 确保日志系统已初始化
	logger := GetLogger()

	logger.Debug("创建新的IoC容器实例")

	return &containerImpl{
		beans:        make(map[string]*BeanDefinition),
		typeRegistry: make(map[string]map[string]*BeanDefinition),
		initializing: make(map[string]bool),
		currentPhase: NotInitialized,
		logger:       logger,
	}
}

// Register 注册一个依赖到容器
func (c *containerImpl) Register(name string, instance interface{}, scope Scope) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		c.logger.Error("容器已初始化，无法注册新的bean",
			zap.String("beanName", name))
		return errors.New("cannot register beans after container initialization")
	}

	if instance == nil {
		c.logger.Error("无法注册空实例",
			zap.String("beanName", name))
		return errors.New("cannot register nil instance")
	}

	// 获取实例的类型
	t := reflect.TypeOf(instance)

	// 创建bean定义
	bean := &BeanDefinition{
		Name:     name,
		Type:     t,
		Instance: instance,
		Scope:    scope,
	}

	// 检查是否已存在同名bean
	if _, exists := c.beans[name]; exists {
		c.logger.Error("bean名称已存在",
			zap.String("beanName", name))
		return fmt.Errorf("bean with name '%s' already exists", name)
	}

	c.beans[name] = bean
	c.logger.Debug("成功注册bean",
		zap.String("beanName", name),
		zap.String("type", t.String()),
		zap.Int("scope", scope))
	return nil
}

// RegisterType 按类型注册依赖
func (c *containerImpl) RegisterType(typeName string, instance interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return errors.New("cannot register beans after container initialization")
	}

	if instance == nil {
		return errors.New("cannot register nil instance")
	}

	// 获取实例的类型
	t := reflect.TypeOf(instance)

	// 获取类型名
	fullTypeName := t.String()

	// 如果是指针，获取指向的元素类型
	if t.Kind() == reflect.Ptr {
		fullTypeName = t.Elem().String()
	}

	// 提取简短类型名（不含包路径）
	parts := strings.Split(fullTypeName, ".")
	shortTypeName := parts[len(parts)-1]

	// 为确保唯一性，生成bean名称
	beanName := shortTypeName
	if typeName != "" {
		beanName = typeName + ":" + shortTypeName
	}

	// 创建bean定义
	bean := &BeanDefinition{
		Name:     beanName,
		TypeName: typeName,
		Type:     t,
		Instance: instance,
		Scope:    Singleton, // 默认为单例
	}

	// 检查是否已存在同名bean
	if _, exists := c.beans[beanName]; exists {
		return fmt.Errorf("bean with name '%s' already exists", beanName)
	}

	// 注册到总表
	c.beans[beanName] = bean

	// 注册到类型表
	if _, exists := c.typeRegistry[typeName]; !exists {
		c.typeRegistry[typeName] = make(map[string]*BeanDefinition)
	}
	c.typeRegistry[typeName][shortTypeName] = bean

	return nil
}

// RegisterTypeWithName 按类型注册依赖,支持配置名字
func (c *containerImpl) RegisterTypeWithName(typeName string, name string, instance interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return errors.New("cannot register beans after container initialization")
	}

	if instance == nil {
		return errors.New("cannot register nil instance")
	}

	// 获取实例的类型
	t := reflect.TypeOf(instance)

	// 如果是指针，获取指向的元素类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 为确保唯一性，生成bean名称
	beanName := typeName + ":" + name

	// 创建bean定义
	bean := &BeanDefinition{
		Name:     beanName,
		TypeName: typeName,
		Type:     t,
		Instance: instance,
		Scope:    Singleton, // 默认为单例
	}

	// 检查是否已存在同名bean
	if _, exists := c.beans[beanName]; exists {
		return fmt.Errorf("bean with name '%s' already exists", beanName)
	}

	// 注册到总表
	c.beans[beanName] = bean

	// 注册到类型表
	if _, exists := c.typeRegistry[typeName]; !exists {
		c.typeRegistry[typeName] = make(map[string]*BeanDefinition)
	}

	// 检查是否已存在同名bean
	if _, exists := c.typeRegistry[typeName][name]; exists {
		return fmt.Errorf("type '%s' already has a bean with name '%s'", typeName, name)
	}

	c.typeRegistry[typeName][name] = bean

	return nil
}

// RegisterFactory 注册一个工厂函数用于创建bean实例
func (c *containerImpl) RegisterFactory(name string, scope Scope, factory func() (interface{}, error)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return errors.New("cannot register beans after container initialization")
	}

	if factory == nil {
		return errors.New("factory function cannot be nil")
	}

	// 创建bean定义
	bean := &BeanDefinition{
		Name:    name,
		Scope:   scope,
		Factory: factory,
	}

	// 检查是否已存在同名bean
	if _, exists := c.beans[name]; exists {
		return fmt.Errorf("bean with name '%s' already exists", name)
	}

	c.beans[name] = bean
	return nil
}

// Get 获取依赖
func (c *containerImpl) Get(name string) interface{} {
	instance, err := c.GetSafe(name)
	if err != nil {
		c.logger.Error("获取bean失败",
			zap.String("beanName", name),
			zap.Error(err))
		panic(err)
	}
	c.logger.Debug("获取bean成功", zap.String("beanName", name))
	return instance
}

// GetSafe 安全地获取依赖，返回错误而不是panic
func (c *containerImpl) GetSafe(name string) (interface{}, error) {
	c.mu.RLock()

	// 获取bean定义
	bean, exists := c.beans[name]
	if !exists {
		c.mu.RUnlock()
		c.logger.Error("找不到bean", zap.String("beanName", name))
		return nil, fmt.Errorf("bean with name '%s' not found", name)
	}

	// 根据当前阶段进行不同处理
	switch c.currentPhase {
	case NotInitialized:
		// 容器尚未初始化
		c.mu.RUnlock()
		c.logger.Error("容器尚未初始化", zap.String("beanName", name))
		return nil, errors.New("container not initialized, call Init() first")

	case InjectionPhase:
		// 注入阶段 - 只返回已存在的实例
		if bean.Instance == nil {
			c.mu.RUnlock()
			// 如果在注入阶段获取尚未初始化的bean，可能导致循环依赖
			if c.initializing[name] {
				c.logger.Error("检测到循环依赖", zap.String("beanName", name))
				return nil, fmt.Errorf("circular dependency detected for bean: %s", name)
			}

			// 尝试初始化该bean
			c.logger.Debug("尝试初始化bean", zap.String("beanName", name))
			c.mu.RUnlock()
			c.mu.Lock()
			defer c.mu.Unlock()

			if err := c.createBeanInstance(name, bean); err != nil {
				c.logger.Error("创建bean实例失败",
					zap.String("beanName", name),
					zap.Error(err))
				return nil, err
			}
			if err := c.injectBeanDependencies(name, bean); err != nil {
				c.logger.Error("注入bean依赖失败",
					zap.String("beanName", name),
					zap.Error(err))
				return nil, err
			}

			c.logger.Debug("bean初始化成功", zap.String("beanName", name))
			return bean.Instance, nil
		}
		instance := bean.Instance
		c.mu.RUnlock()
		return instance, nil

	case PostConstructPhase, Initialized:
		// PostConstruct阶段或已初始化 - 正常返回实例
		if bean.Scope == Singleton {
			instance := bean.Instance
			c.mu.RUnlock()
			return instance, nil
		}

		// 对于Prototype，创建新实例
		c.mu.RUnlock()
		return c.createPrototypeInstance(bean)
	}

	c.mu.RUnlock()
	c.logger.Error("未知的容器状态",
		zap.String("beanName", name),
		zap.Int("phase", c.currentPhase))
	return nil, errors.New("unknown container state")
}

// 创建Prototype类型的bean实例
func (c *containerImpl) createPrototypeInstance(bean *BeanDefinition) (interface{}, error) {
	var instance interface{}
	var err error

	// 使用工厂函数或创建新实例
	if bean.Factory != nil {
		instance, err = bean.Factory()
		if err != nil {
			return nil, err
		}
	} else {
		instance = reflect.New(bean.Type.Elem()).Interface()
	}

	// 注入依赖
	if err := c.Inject(instance); err != nil {
		return nil, err
	}

	// 调用初始化方法
	if initializer, ok := instance.(InitializingBean); ok {
		if err := initializer.PostConstruct(); err != nil {
			return nil, err
		}
	}

	return instance, nil
}

// GetByType 按类型获取依赖
func (c *containerImpl) GetByType(typeName string, name string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 检查容器是否已初始化
	if c.currentPhase == NotInitialized {
		panic("container not initialized, call Init() first")
	}

	// 获取类型注册表
	typeMap, exists := c.typeRegistry[typeName]
	if !exists {
		panic(fmt.Sprintf("no beans registered for type '%s'", typeName))
	}

	// 获取特定名称的bean
	bean, exists := typeMap[name]
	if !exists {
		panic(fmt.Sprintf("bean with name '%s' of type '%s' not found", name, typeName))
	}

	// 单例直接返回实例
	if bean.Scope == Singleton {
		return bean.Instance
	}

	// 多例创建新实例（简化处理，完整实现应该像Get方法一样）
	newInstance := reflect.New(bean.Type.Elem()).Interface()
	return newInstance
}

// GetAll 获取所有注册的bean
func (c *containerImpl) GetAll() map[string]*BeanDefinition {
	return c.beans
}

// GetAllNames 获取所有注册的bean名称
func (c *containerImpl) GetAllNames() []string {
	names := make([]string, 0, len(c.beans))
	for name := range c.beans {
		names = append(names, name)
	}
	return names
}

// Inject 注入依赖
func (c *containerImpl) Inject(instance interface{}) error {
	if instance == nil {
		return errors.New("cannot inject into nil instance")
	}

	val := reflect.ValueOf(instance)

	// 如果是指针，获取其元素
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 只能向结构体注入
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("can only inject into struct, got %s", val.Kind())
	}

	// 获取类型
	t := val.Type()

	// 遍历所有字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 检查是否有注入标签
		injectTag := field.Tag.Get("inject")
		if injectTag == "" {
			continue
		}

		// 检查是否为可选依赖
		optionalTag := field.Tag.Get("optional")
		optional := optionalTag == "true"

		// 获取要注入的bean
		var bean interface{}
		var err error

		if injectTag == "" {
			// 自动查找匹配类型的bean
			bean, err = c.findCandidateByType(field.Type)
		} else {
			// 根据名称获取bean
			bean, err = c.GetSafe(injectTag)
		}

		// 处理查找/获取错误
		if err != nil {
			if optional {
				// 可选依赖，跳过注入
				continue
			}
			return fmt.Errorf("error injecting field '%s': %w", field.Name, err)
		}

		// 设置字段值
		fieldVal := val.Field(i)
		if !fieldVal.CanSet() {
			return fmt.Errorf("cannot set field '%s', it might be unexported", field.Name)
		}

		beanVal := reflect.ValueOf(bean)

		// 处理接口类型的注入
		if field.Type.Kind() == reflect.Interface && !beanVal.Type().Implements(field.Type) {
			return fmt.Errorf("bean of type %s does not implement interface %s", beanVal.Type(), field.Type)
		}

		fieldVal.Set(beanVal)
	}

	return nil
}

// 查找匹配类型的bean候选
func (c *containerImpl) findCandidateByType(t reflect.Type) (interface{}, error) {
	var candidates []string

	for name, bean := range c.beans {
		beanType := bean.Type

		// 检查类型匹配
		if t.Kind() == reflect.Interface {
			// 如果要注入的是接口，检查bean是否实现了该接口
			if beanType.Implements(t) {
				candidates = append(candidates, name)
			}
		} else if t == beanType {
			// 直接的类型匹配
			candidates = append(candidates, name)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no bean candidate found for type %s", t)
	}

	if len(candidates) > 1 {
		return nil, fmt.Errorf("multiple bean candidates found for type %s: %v", t, candidates)
	}

	// 获取唯一的候选bean
	return c.Get(candidates[0]), nil
}

// Init 实现两阶段初始化
func (c *containerImpl) Init() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		c.logger.Warn("容器已经初始化，无需再次初始化")
		return errors.New("container already initialized")
	}

	c.logger.Info("开始初始化IoC容器", zap.Int("beanCount", len(c.beans)))

	// 第一阶段：依赖注入
	c.logger.Info("进入第一阶段：依赖注入")
	c.currentPhase = InjectionPhase

	// 先为所有单例bean创建实例
	c.logger.Debug("开始创建所有单例bean实例")
	for name, bean := range c.beans {
		if bean.Scope == Singleton {
			if err := c.createBeanInstance(name, bean); err != nil {
				c.logger.Error("创建bean实例失败",
					zap.String("beanName", name),
					zap.Error(err))
				return err
			}
		}
	}

	// 然后为所有单例bean注入依赖
	c.logger.Debug("开始为所有单例bean注入依赖")
	for name, bean := range c.beans {
		if bean.Scope == Singleton {
			if err := c.injectBeanDependencies(name, bean); err != nil {
				c.logger.Error("注入bean依赖失败",
					zap.String("beanName", name),
					zap.Error(err))
				return err
			}
		}
	}

	// 第二阶段：调用所有PostConstruct方法
	c.logger.Info("进入第二阶段：初始化")
	c.currentPhase = PostConstructPhase
	for name, bean := range c.beans {
		if bean.Scope == Singleton {
			if err := c.initializeBean(name, bean); err != nil {
				c.logger.Error("初始化bean失败",
					zap.String("beanName", name),
					zap.Error(err))
				return err
			}
		}
	}

	// 标记初始化完成
	c.initialized = true
	c.currentPhase = Initialized
	c.logger.Info("IoC容器初始化完成",
		zap.Int("beanCount", len(c.beans)))
	return nil
}

// 创建bean实例
func (c *containerImpl) createBeanInstance(name string, bean *BeanDefinition) error {
	// 如果已经有实例，跳过
	if bean.Instance != nil {
		c.logger.Debug("bean实例已存在，跳过创建",
			zap.String("beanName", name))
		return nil
	}

	// 检测循环依赖
	if c.initializing[name] {
		c.logger.Error("检测到循环依赖",
			zap.String("beanName", name))
		return fmt.Errorf("circular dependency detected for bean: %s", name)
	}

	c.logger.Debug("开始创建bean实例",
		zap.String("beanName", name))
	c.initializing[name] = true
	defer delete(c.initializing, name)

	// 如果是工厂方法，调用它创建实例
	if bean.Factory != nil {
		c.logger.Debug("使用工厂方法创建bean实例",
			zap.String("beanName", name))
		instance, err := bean.Factory()
		if err != nil {
			c.logger.Error("工厂方法创建实例失败",
				zap.String("beanName", name),
				zap.Error(err))
			return fmt.Errorf("error creating instance for bean '%s': %w", name, err)
		}
		bean.Instance = instance

		// 获取实例的类型
		t := reflect.TypeOf(instance)
		bean.Type = t
		c.logger.Debug("成功创建bean实例",
			zap.String("beanName", name),
			zap.String("type", t.String()))
	}

	return nil
}

// 注入bean依赖
func (c *containerImpl) injectBeanDependencies(name string, bean *BeanDefinition) error {
	// 如果已经注入过，跳过
	if bean.injected {
		c.logger.Debug("bean已注入依赖，跳过注入",
			zap.String("beanName", name))
		return nil
	}

	c.logger.Debug("开始为bean注入依赖",
		zap.String("beanName", name))

	// 临时释放锁，避免在注入过程中发生死锁
	c.mu.Unlock()

	// 注入依赖
	err := c.injectDuringInit(bean.Instance)

	// 重新获取锁
	c.mu.Lock()

	if err != nil {
		c.logger.Error("注入依赖失败",
			zap.String("beanName", name),
			zap.Error(err))
		return fmt.Errorf("error injecting dependencies for bean '%s': %w", name, err)
	}

	// 标记为已注入
	bean.injected = true
	c.logger.Debug("成功为bean注入依赖",
		zap.String("beanName", name))
	return nil
}

// 初始化bean
func (c *containerImpl) initializeBean(name string, bean *BeanDefinition) error {
	// 如果已经初始化过，跳过
	if bean.initialized {
		c.logger.Debug("bean已初始化，跳过",
			zap.String("beanName", name))
		return nil
	}

	// 确保已经注入了依赖
	if !bean.injected {
		c.logger.Debug("bean尚未注入依赖，先注入依赖",
			zap.String("beanName", name))
		if err := c.injectBeanDependencies(name, bean); err != nil {
			return err
		}
	}

	// 检查是否实现了InitializingBean接口
	initializer, ok := bean.Instance.(InitializingBean)
	if !ok {
		// 未实现接口，标记为已初始化并返回
		c.logger.Debug("bean未实现InitializingBean接口，跳过PostConstruct",
			zap.String("beanName", name))
		bean.initialized = true
		return nil
	}

	c.logger.Debug("调用bean的PostConstruct方法",
		zap.String("beanName", name))

	// 临时释放锁，避免在PostConstruct过程中发生死锁
	c.mu.Unlock()

	// 调用PostConstruct方法
	err := initializer.PostConstruct()

	// 重新获取锁
	c.mu.Lock()

	if err != nil {
		c.logger.Error("PostConstruct方法执行失败",
			zap.String("beanName", name),
			zap.Error(err))
		return fmt.Errorf("error initializing bean '%s': %w", name, err)
	}

	// 标记为已初始化
	bean.initialized = true
	c.logger.Debug("bean初始化完成",
		zap.String("beanName", name))
	return nil
}

// injectDuringInit 在初始化过程中注入依赖，不使用容器的锁
func (c *containerImpl) injectDuringInit(instance interface{}) error {
	if instance == nil {
		return errors.New("cannot inject into nil instance")
	}

	val := reflect.ValueOf(instance)

	// 如果是指针，获取其元素
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 只能向结构体注入
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("can only inject into struct, got %s", val.Kind())
	}

	// 获取类型
	t := val.Type()

	// 遍历所有字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 检查是否有注入标签
		injectTag := field.Tag.Get("inject")
		if injectTag == "" {
			continue
		}

		// 检查是否为可选依赖
		optionalTag := field.Tag.Get("optional")
		optional := optionalTag == "true"

		// 获取要注入的bean
		var bean interface{}
		var err error

		// 在初始化过程中，需要手动查找bean而不是使用GetSafe
		if injectTag != "" {
			c.mu.RLock()
			beanDef, exists := c.beans[injectTag]
			c.mu.RUnlock()

			if !exists {
				if optional {
					continue
				}
				return fmt.Errorf("bean with name '%s' not found", injectTag)
			}

			// 如果bean实例尚未创建，则创建它
			if beanDef.Instance == nil {
				c.mu.Lock()
				if err := c.createBeanInstance(injectTag, beanDef); err != nil {
					c.mu.Unlock()
					if optional {
						continue
					}
					return err
				}
				c.mu.Unlock()
			}

			bean = beanDef.Instance
		} else {
			// 自动查找匹配类型的bean候选
			bean, err = c.findCandidateByTypeForInit(field.Type)
			if err != nil {
				if optional {
					continue
				}
				return fmt.Errorf("error injecting field '%s': %w", field.Name, err)
			}
		}

		// 设置字段值
		fieldVal := val.Field(i)
		if !fieldVal.CanSet() {
			return fmt.Errorf("cannot set field '%s', it might be unexported", field.Name)
		}

		beanVal := reflect.ValueOf(bean)

		// 处理接口类型的注入
		if field.Type.Kind() == reflect.Interface && !beanVal.Type().Implements(field.Type) {
			return fmt.Errorf("bean of type %s does not implement interface %s", beanVal.Type(), field.Type)
		}

		fieldVal.Set(beanVal)
	}

	return nil
}

// findCandidateByTypeForInit 在初始化过程中查找匹配类型的bean候选，不使用容器的锁获取bean
func (c *containerImpl) findCandidateByTypeForInit(t reflect.Type) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var candidates []string
	var candidateInstances []interface{}

	for name, bean := range c.beans {
		if bean.Instance == nil {
			continue // 跳过尚未初始化的bean
		}

		beanType := bean.Type

		// 检查类型匹配
		if t.Kind() == reflect.Interface {
			// 如果要注入的是接口，检查bean是否实现了该接口
			if beanType != nil && beanType.Implements(t) {
				candidates = append(candidates, name)
				candidateInstances = append(candidateInstances, bean.Instance)
			}
		} else if t == beanType {
			// 直接的类型匹配
			candidates = append(candidates, name)
			candidateInstances = append(candidateInstances, bean.Instance)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no bean candidate found for type %s", t)
	}

	if len(candidates) > 1 {
		return nil, fmt.Errorf("multiple bean candidates found for type %s: %v", t, candidates)
	}

	// 返回唯一的候选bean实例
	return candidateInstances[0], nil
}
