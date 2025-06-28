package model

type ProductsResult struct {
	Products []ProductDTO `json:"products"`
	Total    uint         `json:"total" db:"total"`
}
type ProductDTO struct {
	ID            uint64  `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	OriginalPrice float64 `json:"originalPrice" db:"price"`
	Price         float64 `json:"price"`
	Category      string  `json:"category"`
	CategoryName  string  `json:"categoryName" db:"category_name"`
	Stock         float64 `json:"stock" db:"stock"`
}

type ProductDetailDTO struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"categoryName"`
	DefaultCode  string  `json:"default_code"`
	Description  string  `json:"description"`
}

type BestSellerDTO struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	QtySold     float64 `json:"qty_sold"`
}

type Admin struct {
	ID       int64  `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}
