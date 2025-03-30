package product

// ProductService 产品服务接口
type ProductService interface {
	GetProduct(id string) (*Product, error)
	ListProducts() ([]*Product, error)
	CreateProduct(product *Product) error
	UpdateProduct(product *Product) error
	DeleteProduct(id string) error
}
