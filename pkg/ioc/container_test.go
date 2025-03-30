package ioc_test

import (
	"testing"

	"github.com/TickleLee/ioc/pkg/ioc"
)

type ProductService interface {
	GetProduct(id string) string
}

type ProductServiceImpl struct{}

func (p *ProductServiceImpl) GetProduct(id string) string {
	return id
}

type QuotaService interface {
	GetQuota(id string) string
}

type QuotaServiceImpl struct {
	ProductService ProductService `inject:"productService"`
}

func (q *QuotaServiceImpl) GetQuota(id string) string {
	return q.ProductService.GetProduct(id)
}

func TestContainer(t *testing.T) {

	var productService ProductService
	container := ioc.NewContainer()
	container.Register("productService", &ProductServiceImpl{}, ioc.Singleton)

	container.Init()

	productService = container.Get("productService").(ProductService)

	t.Log(productService.GetProduct("123"))
}

func TestContainer_RegisterType(t *testing.T) {
	container := ioc.NewContainer()
	container.RegisterType("Service", &ProductServiceImpl{})

	container.Init()

	productService := container.GetByType("Service", "ProductServiceImpl").(ProductService)

	t.Log(productService.GetProduct("123"))
}

func TestContainer_RegisterType_BYGet(t *testing.T) {
	container := ioc.NewContainer()
	container.RegisterType("Service", &ProductServiceImpl{})

	container.Init()

	productService := container.Get("Service:ProductServiceImpl").(ProductService)

	t.Log(productService.GetProduct("123"))
}

func TestContainer_RegisterFactory(t *testing.T) {
	container := ioc.NewContainer()
	container.RegisterFactory("productService", ioc.Singleton, func() (interface{}, error) {
		return &ProductServiceImpl{}, nil
	})

	container.Init()

	productService := container.Get("productService").(ProductService)

	t.Log(productService.GetProduct("123"))
}

func TestContainer_GetSafe(t *testing.T) {
	container := ioc.NewContainer()
	container.Register("productService", &ProductServiceImpl{}, ioc.Singleton)

	container.Init()

	productService, err := container.GetSafe("productService")
	if err != nil {
		t.Fatal(err)
	}

	productServiceImpl := productService.(ProductService)

	t.Log(productServiceImpl.GetProduct("123"))
}

// 测试一个错误的bean名称
func TestContainer_GetSafe_BYGet(t *testing.T) {
	container := ioc.NewContainer()
	container.Register("productService", &ProductServiceImpl{}, ioc.Singleton)

	container.Init()

	_, err := container.GetSafe("productService2")
	if err != nil {
		t.Fatal(err)
	}
	// 并不会panic
}

func TestContainer_Inject(t *testing.T) {
	container := ioc.NewContainer()

	container.Register("productService", &ProductServiceImpl{}, ioc.Singleton)
	container.Register("quotaService", &QuotaServiceImpl{}, ioc.Singleton)

	container.Init()

	container.Inject(&QuotaServiceImpl{})

	quotaService := container.Get("quotaService").(QuotaService)

	t.Log(quotaService.GetQuota("123"))
}

func TestContainer_Inject_Optional(t *testing.T) {
	container := ioc.NewContainer()

	container.Register("productService", &ProductServiceImpl{}, ioc.Singleton)
	container.Register("quotaService", &QuotaServiceImpl{}, ioc.Singleton)

	container.Inject(&QuotaServiceImpl{})

	quotaService := container.Get("quotaService").(QuotaService)

	if quotaService == nil {
		t.Fatal("quotaService is nil")
	}

	t.Log(quotaService.GetQuota("123"))
}

func TestContainer_GetAll(t *testing.T) {
	container := ioc.NewContainer()

	container.Register("productService", &ProductServiceImpl{}, ioc.Singleton)
	container.Register("quotaService", &QuotaServiceImpl{}, ioc.Singleton)

	container.Init()

	beans := container.GetAll()

	for _, bean := range beans {
		t.Log(bean.Name, bean.Type)
	}
}
