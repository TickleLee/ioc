package logger

import (
	"fmt"
	"log"
	"time"

	"github.com/TickleLee/ioc/examples/examples/config"
	"github.com/TickleLee/ioc/pkg/ioc"
)

// LogService 日志服务接口
type LogService interface {
	Info(message string)
	Error(message string)
	Debug(message string)
}

// LogServiceImpl 日志服务实现
type LogServiceImpl struct {
	Config *config.Config `inject:"appConfig"`
}

// PostConstruct 初始化方法
func (l *LogServiceImpl) PostConstruct() error {
	fmt.Println("初始化 LogService，日志级别:", l.Config.LogLevel)
	return nil
}

func (l *LogServiceImpl) Info(message string) {
	fmt.Printf("[INFO] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

func (l *LogServiceImpl) Error(message string) {
	fmt.Printf("[ERROR] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

func (l *LogServiceImpl) Debug(message string) {
	if l.Config.LogLevel == "DEBUG" {
		fmt.Printf("[DEBUG] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
	}
}

func init() {
	err := ioc.Register("logService", &LogServiceImpl{}, ioc.Singleton)
	if err != nil {
		log.Fatalf("注册 logService 失败: %v", err)
	}
}
