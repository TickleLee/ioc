package product

import (
	"fmt"
	"log"

	"github.com/TickleLee/ioc/pkg/ioc"
)

type ProductController struct {
	ProductService ProductService `inject:"productService"`
}

func (c *ProductController) ShowProduct(id string) {
	product, err := c.ProductService.GetProduct(id)
	if err != nil {
		fmt.Printf("获取产品失败: %v\n", err)
		return
	}

	fmt.Printf("产品详情: ID=%s, 名称=%s, 描述=%s, 价格=%.2f\n",
		product.ID, product.Name, product.Description, product.Price)
}

func (c *ProductController) CreateNewProduct(id, name, desc string, price float64) {
	product := &Product{
		ID:          id,
		Name:        name,
		Description: desc,
		Price:       price,
	}

	err := c.ProductService.CreateProduct(product)
	if err != nil {
		fmt.Printf("创建产品失败: %v\n", err)
		return
	}

	fmt.Println("创建产品成功!")
}

func init() {
	err := ioc.Register("productController", &ProductController{}, ioc.Singleton)
	if err != nil {
		log.Fatalf("注册 productController 失败: %v", err)
	}
}
