package impl

import (
	"errors"
	"fmt"
	"log"

	"github.com/TickleLee/ioc/examples/examples/config"
	"github.com/TickleLee/ioc/examples/examples/modules/product"
	"github.com/TickleLee/ioc/pkg/ioc"
)

// ProductRepositoryImpl 产品仓储实现
type ProductRepositoryImpl struct {
	// 模拟数据库
	products map[string]*product.Product
	Config   *config.Config `inject:"appConfig"`
}

// PostConstruct 初始化方法
func (r *ProductRepositoryImpl) PostConstruct() error {
	// 初始化仓储，模拟数据库连接
	fmt.Println("初始化 ProductRepository，使用配置:", r.Config.DatabaseURL)
	r.products = make(map[string]*product.Product)

	// 添加一些示例数据
	r.products["p001"] = &product.Product{ID: "p001", Name: "笔记本电脑", Description: "高性能笔记本", Price: 5999.00}
	r.products["p002"] = &product.Product{ID: "p002", Name: "智能手机", Description: "新一代智能手机", Price: 3999.00}

	return nil
}

func (r *ProductRepositoryImpl) FindByID(id string) (*product.Product, error) {
	if product, exists := r.products[id]; exists {
		return product, nil
	}
	return nil, errors.New("产品不存在")
}

func (r *ProductRepositoryImpl) FindAll() ([]*product.Product, error) {
	products := make([]*product.Product, 0, len(r.products))
	for _, product := range r.products {
		products = append(products, product)
	}
	return products, nil
}

func (r *ProductRepositoryImpl) Save(product *product.Product) error {
	if product.ID == "" {
		return errors.New("产品ID不能为空")
	}
	r.products[product.ID] = product
	return nil
}

func (r *ProductRepositoryImpl) Delete(id string) error {
	if _, exists := r.products[id]; !exists {
		return errors.New("产品不存在")
	}
	delete(r.products, id)
	return nil
}

func init() {
	err := ioc.Register("productRepository", &ProductRepositoryImpl{}, ioc.Singleton)
	if err != nil {
		log.Fatalf("注册 productRepository 失败: %v", err)
	}
}
