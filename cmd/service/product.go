package service

import (
	"fmt"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/model"
	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/repository"
)

type ProductService interface {
	GetAll(page, pageSize int) (*model.ProductsResult, error)
	GetByID(id int64) (*model.ProductDTO, error)
	GetFiltered(page, pageSize int, categID, minPrice, maxPrice *int64, category []string, name, orderValue string) (*model.ProductsResult, error)
	GetRelated(category, name string, page, page_size int) (*model.ProductsResult, error)
	GetBestSelling(limit int) ([]model.ProductDTO, error)
	GetVariants(productID int64) ([]model.ProductDTO, error)
	GetCategorys() ([]repository.Category, error)
}

type productService struct {
	repo repository.ProductRepo
}

// NewProductService construye el servicio a partir de un ProductRepo.
func NewProductService(r repository.ProductRepo) ProductService {
	return &productService{repo: r}
}

// GetAll aplica paginación a partir de page y pageSize.
func (s *productService) GetAll(page, pageSize int) (*model.ProductsResult, error) {
	if page < 1 {
		return nil, fmt.Errorf("page debe ser >= 1")
	}
	offset := (page - 1) * pageSize
	return s.repo.GetAll(offset, pageSize)
}
func (s *productService) GetByID(id int64) (*model.ProductDTO, error) {

	product, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("error al obtener producto: %w", err)
	}
	return product, nil
}

// GetFiltered delega el filtrado con paginación al repo.
func (s *productService) GetFiltered(page, pageSize int, categID, minPrice, maxPrice *int64, category []string, name, orderValue string) (*model.ProductsResult, error) {
	if page < 1 {
		return nil, fmt.Errorf("page debe ser >= 1")
	}
	offset := (page - 1) * pageSize
	return s.repo.GetFiltered(offset, pageSize, categID, minPrice, maxPrice, category, name, orderValue)
}

// GetRelated toma el límite y delega a repo.
func (s *productService) GetRelated(category, name string, page, page_size int) (*model.ProductsResult, error) {

	if page_size < 1 {
		page_size = 5
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * page_size

	return s.repo.GetRelated(category, name, &offset, &page_size)
}

// GetBestSelling delega a repo (limit por defecto si se pasa 0).
func (s *productService) GetBestSelling(limit int) ([]model.ProductDTO, error) {
	if limit < 1 {
		limit = 6
	}
	return s.repo.GetBestSelling(limit)
}
func (s *productService) GetCategorys() ([]repository.Category, error) {
	return s.repo.GetCategorys()
}

// GetVariants delega a repo.
func (s *productService) GetVariants(productID int64) ([]model.ProductDTO, error) {
	if productID <= 0 {
		return nil, fmt.Errorf("productID inválido")
	}
	return s.repo.GetVariants(productID)
}
