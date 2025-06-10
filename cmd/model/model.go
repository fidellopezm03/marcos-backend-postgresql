package model

type ProductsResult struct {
	Products []ProductDTO `json:"products"`
	Total    uint         `json:"total" db:"total"`
}
type ProductDTO struct {
	ID           uint64  `json:"id" db:"id"`
	Name         string  `json:"name" db:"name"`
	Price        float64 `json:"price" db:"price"`
	Category     string  `json:"category"`
	CategoryName string  `json:"category_name" db:"category_name"`
	Stock        float64 `json:"stock" db:"stock"`
}

type ProductDetailDTO struct {
	ID           int64   `json:"id" db:""`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	DefaultCode  string  `json:"default_code"`
	Description  string  `json:"description"`
}

type BestSellerDTO struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	QtySold     float64 `json:"qty_sold"`
}
