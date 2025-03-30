package config

import (
	"log"
	"time"

	"github.com/TickleLee/ioc/pkg/ioc"
)

// Config 配置结构体
type Config struct {
	LogLevel       string
	DatabaseURL    string
	QuotaResetTime time.Duration
	MaxQuota       int
}

func init() {
	// 注册配置
	config := &Config{
		LogLevel:       "DEBUG",
		DatabaseURL:    "mysql://localhost:3306/productdb",
		QuotaResetTime: 1 * time.Hour,
		MaxQuota:       5,
	}

	err := ioc.Register("appConfig", config, ioc.Singleton)
	if err != nil {
		log.Fatalf("注册 appConfig 失败: %v", err)
	}
}
