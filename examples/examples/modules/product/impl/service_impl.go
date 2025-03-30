package impl

import (
	"errors"
	"fmt"
	"log"

	"github.com/TickleLee/ioc/examples/examples/logger"
	"github.com/TickleLee/ioc/examples/examples/modules/product"
	"github.com/TickleLee/ioc/examples/examples/modules/quota"
	"github.com/TickleLee/ioc/pkg/ioc"
)

// ProductServiceImpl 产品服务实现
type ProductServiceImpl struct {
	Repo     product.ProductRepository `inject:"productRepository"`
	QuotaSvc quota.QuotaService        `inject:"quotaService"`
	Logger   logger.LogService         `inject:"logService"`
}

func (s *ProductServiceImpl) PostConstruct() error {
	fmt.Println("初始化 ProductService")
	return nil
}

func (s *ProductServiceImpl) GetProduct(id string) (*product.Product, error) {
	s.Logger.Info("获取产品:" + id)
	return s.Repo.FindByID(id)
}

func (s *ProductServiceImpl) ListProducts() ([]*product.Product, error) {
	s.Logger.Info("获取所有产品")
	return s.Repo.FindAll()
}

func (s *ProductServiceImpl) CreateProduct(product *product.Product) error {
	s.Logger.Info("创建产品:" + product.ID)

	// 检查配额
	if !s.QuotaSvc.HasQuota("createProduct") {
		return errors.New("创建产品的配额已用完，请稍后再试")
	}

	// 消耗配额
	s.QuotaSvc.UseQuota("createProduct")

	return s.Repo.Save(product)
}

func (s *ProductServiceImpl) UpdateProduct(product *product.Product) error {
	s.Logger.Info("更新产品:" + product.ID)

	// 先检查产品是否存在
	_, err := s.Repo.FindByID(product.ID)
	if err != nil {
		return err
	}

	return s.Repo.Save(product)
}

func (s *ProductServiceImpl) DeleteProduct(id string) error {
	s.Logger.Info("删除产品:" + id)
	return s.Repo.Delete(id)
}

func init() {
	err := ioc.Register("productService", &ProductServiceImpl{}, ioc.Singleton)
	if err != nil {
		log.Fatalf("注册 productService 失败: %v", err)
	}
}
