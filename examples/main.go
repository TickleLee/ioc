package main

import (
	"fmt"
	"log"

	"github.com/TickleLee/ioc/examples/examples/modules/product"
	"github.com/TickleLee/ioc/pkg/ioc"

	// 注册配置模块
	_ "github.com/TickleLee/ioc/examples/examples/config"

	// 注册日志模块
	_ "github.com/TickleLee/ioc/examples/examples/logger"

	// 注册产品模块
	_ "github.com/TickleLee/ioc/examples/examples/modules/product/impl"

	// 注册配额模块
	_ "github.com/TickleLee/ioc/examples/examples/modules/quota/impl"
)

func main() {
	fmt.Println("====================================")
	fmt.Println("启动 IoC 产品模块和配额管理示例")
	fmt.Println("====================================")

	// 运行产品模块示例
	RunProductExample()

	fmt.Println("\n====================================")
	fmt.Println("示例运行结束")
	fmt.Println("====================================")
}

func RunProductExample() {
	// 初始化容器
	err := ioc.Init()
	if err != nil {
		log.Fatalf("初始化容器失败: %v", err)
	}

	fmt.Println("\n===== 产品和配额管理系统 =====")

	// 获取产品控制器
	controller := ioc.Get("productController").(*product.ProductController)

	// 显示现有产品
	controller.ShowProduct("p001")

	// 创建新产品
	controller.CreateNewProduct("p003", "平板电脑", "10英寸高清平板", 2499.00)

	// 演示手动注入
	fmt.Println("\n===== 演示手动注入 =====")
	manualExample()

	// 演示配额限制
	fmt.Println("\n===== 演示配额限制 =====")
	quotaLimitExample()
}

// 手动注入示例
func manualExample() {
	// 创建一个全新的产品控制器，未注入依赖
	customController := &product.ProductController{}

	// 此时直接使用会导致空指针错误
	fmt.Println("注入前，ProductService为nil:", customController.ProductService == nil)

	// 手动注入依赖
	err := ioc.Inject(customController)
	if err != nil {
		fmt.Printf("手动注入失败: %v\n", err)
		return
	}

	// 注入后可以正常使用
	fmt.Println("注入后，ProductService不为nil:", customController.ProductService != nil)

	// 使用控制器
	customController.ShowProduct("p002")
}

// 配额限制示例
func quotaLimitExample() {
	controller := ioc.Get("productController").(*product.ProductController)

	// 重复创建产品，直到超出配额
	for i := 1; i <= 6; i++ {
		id := fmt.Sprintf("p%03d", i+100)
		controller.CreateNewProduct(id, fmt.Sprintf("测试产品%d", i), "配额测试", 999.00)
	}
}
