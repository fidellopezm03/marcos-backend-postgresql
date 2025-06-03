package service

import (
	"fmt"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/model"
	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/repository"
)

type ProductService interface {
    GetAll(page, pageSize int) ([]model.ProductDTO, error)
    GetFiltered(page, pageSize int, categID, minPrice, maxPrice *int64) ([]model.ProductDTO, error)
    GetRelated(productID int64, limit int) ([]model.ProductDTO, error)
    GetBestSelling(limit int) ([]model.ProductDTO, error)
    GetVariants(productID int64) ([]model.ProductDTO, error)
}

type productService struct {
    repo repository.ProductRepo
}

// NewProductService construye el servicio a partir de un ProductRepo.
func NewProductService(r repository.ProductRepo) ProductService {
    return &productService{repo: r}
}

// GetAll aplica paginación a partir de page y pageSize.
func (s *productService) GetAll(page, pageSize int) ([]model.ProductDTO, error) {
    if page < 1 {
        return nil, fmt.Errorf("page debe ser >= 1")
    }
    offset := (page - 1) * pageSize
    return s.repo.GetAll(offset, pageSize)
}

// GetFiltered delega el filtrado con paginación al repo.
func (s *productService) GetFiltered(page, pageSize int, categID, minPrice, maxPrice *int64) ([]model.ProductDTO, error) {
    if page < 1 {
        return nil, fmt.Errorf("page debe ser >= 1")
    }
    offset := (page - 1) * pageSize
    return s.repo.GetFiltered(offset, pageSize, categID, minPrice, maxPrice)
}

// GetRelated toma el límite y delega a repo.
func (s *productService) GetRelated(productID int64, limit int) ([]model.ProductDTO, error) {
    if productID <= 0 {
        return nil, fmt.Errorf("productID inválido")
    }
    if limit < 1 {
        limit = 5
    }
    return s.repo.GetRelated(productID, limit)
}

// GetBestSelling delega a repo (limit por defecto si se pasa 0).
func (s *productService) GetBestSelling(limit int) ([]model.ProductDTO, error) {
    if limit < 1 {
        limit = 10
    }
    return s.repo.GetBestSelling(limit)
}

// GetVariants delega a repo.
func (s *productService) GetVariants(productID int64) ([]model.ProductDTO, error) {
    if productID <= 0 {
        return nil, fmt.Errorf("productID inválido")
    }
    return s.repo.GetVariants(productID)
}