package main

import (
	"fmt"

	"github.com/TickleLee/ioc/pkg/ioc"
	"go.uber.org/zap"
)

// 定义服务接口
type ExampleService interface {
	DoSomething() error
}

// 服务实现
type ExampleServiceImpl struct {
	Name string
}

func (s *ExampleServiceImpl) DoSomething() error {
	// 获取IoC容器的日志记录器
	logger := ioc.GetContainerLogger()
	logger.Info("服务正在执行操作",
		zap.String("serviceName", s.Name))

	return nil
}

// PostConstruct实现
func (s *ExampleServiceImpl) PostConstruct() error {
	logger := ioc.GetContainerLogger()
	logger.Info("ExampleService初始化",
		zap.String("serviceName", s.Name))
	return nil
}

func main() {
	// 配置日志系统 - 请在任何IoC操作之前调用
	ioc.ConfigureLogging(ioc.LoggerConfig{
		Level:         ioc.DebugLevel,   // 开启调试级别，显示详细日志
		EnableJSON:    false,            // 使用文本格式，方便阅读
		OutputFile:    true,             // 同时输出到文件
		FilePath:      "./logs/ioc.log", // 日志文件路径
		OutputConsole: true,             // 输出到控制台
		EnableCaller:  true,             // 显示调用者信息
		Development:   true,             // 开发模式，显示彩色日志
	})

	// 也可以使用简化的调试模式配置
	// ioc.EnableDebugLogging()

	// 获取日志记录器
	logger := ioc.GetContainerLogger()
	logger.Info("开始演示IoC容器日志系统")

	// 注册服务
	logger.Debug("注册示例服务")
	err := ioc.Register("exampleService", &ExampleServiceImpl{Name: "示例服务1"}, ioc.Singleton)
	if err != nil {
		logger.Error("注册服务失败", zap.Error(err))
		return
	}

	// 注册另一个服务
	logger.Debug("注册另一个示例服务")
	err = ioc.Register("anotherService", &ExampleServiceImpl{Name: "示例服务2"}, ioc.Singleton)
	if err != nil {
		logger.Error("注册服务失败", zap.Error(err))
		return
	}

	// 初始化容器
	logger.Info("初始化IoC容器")
	err = ioc.Init()
	if err != nil {
		logger.Error("初始化容器失败", zap.Error(err))
		return
	}

	// 获取服务
	logger.Debug("获取服务")
	service := ioc.Get("exampleService").(ExampleService)

	// 调用服务方法
	logger.Debug("调用服务方法")
	err = service.DoSomething()
	if err != nil {
		logger.Error("服务方法执行失败", zap.Error(err))
		return
	}

	// 获取并调用另一个服务
	anotherService := ioc.Get("anotherService").(ExampleService)
	err = anotherService.DoSomething()
	if err != nil {
		logger.Error("另一个服务执行失败", zap.Error(err))
		return
	}

	// 测试不同日志级别
	logger.Debug("这是一条调试信息", zap.Int("debugValue", 100))
	logger.Info("这是一条普通信息", zap.String("infoKey", "infoValue"))
	logger.Warn("这是一条警告信息", zap.Bool("warning", true))

	// 使用字段创建新的日志记录器
	contextLogger := logger.With(
		zap.String("component", "LoggingExample"),
		zap.String("context", "示例上下文"),
	)
	contextLogger.Info("这条日志包含上下文信息")

	fmt.Println("示例执行完成，请查看控制台输出和日志文件")
}
