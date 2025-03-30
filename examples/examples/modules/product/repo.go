package product

// ProductRepository 产品仓储接口
type ProductRepository interface {
	FindByID(id string) (*Product, error)
	FindAll() ([]*Product, error)
	Save(product *Product) error
	Delete(id string) error
}
