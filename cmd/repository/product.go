package repository

import (
	"database/sql"
	"fmt"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/model"
)

// ProductRepo define la interfaz para acceso a productos en Odoo.
type ProductRepo interface {
	GetAll(offset, limit int) ([]model.ProductDTO, error)
	GetFiltered(offset, limit int, categID, minPrice, maxPrice *int64) ([]model.ProductDTO, error)
	GetRelated(productID int64, limit int) ([]model.ProductDTO, error)
	GetBestSelling(limit int) ([]model.ProductDTO, error)
	GetVariants(productID int64) ([]model.ProductDTO, error)
}

// odooProductRepo es la implementación concreta que usa go-odoo internamente.
type odooProductRepo struct {
	DB *sql.DB
}

// NewProductRepo construye un repository con un cliente Odoo ya iniciado.
func NewProductRepo(d *sql.DB) ProductRepo {
	return &odooProductRepo{DB: d}
}

// GetAll recupera todos los productos (product.product) con paginación.
func (r *odooProductRepo) GetAll(offset, limit int) ([]model.ProductDTO, error) {
	return r.GetFiltered(offset,limit, nil, nil, nil)}

// GetRelated busca productos relacionados al productID dado.
func(r*odooProductRepo)GetRelated(productID int64, limit int) ([]model.ProductDTO, error) {
	
	return []model.ProductDTO{}, nil;
}
// GetFiltered permite filtrar por categoría (categ_id) y rango de precio list_price.
// Ambos filtros son opcionales: pasar nil para omitir.
func (r *odooProductRepo) GetFiltered(offset, limit int, categID, minPrice, maxPrice *int64) ([]model.ProductDTO, error) {
	var Products []model.ProductDTO
	where := ""
	if minPrice != nil || maxPrice != nil {
		limitPrice := fmt.Sprintf("list_price >= %v and list_price <= %v",minPrice,maxPrice)
		where = " WHERE" + " " + limitPrice + " "
	}
	//const id_location = "8"
	exist := "WITH exist AS (SELECT product_id, SUM(quantity) as stock FROM stock_quant WHERE location_id = 8 GROUP BY product_id HAVING SUM(quantity) > 0 OFFSET $1 LIMIT $2)"
	query := exist + " " + fmt.Sprintf("SELECT product_id as id, name, list_price as price, stock FROM (SELECT product_id, stock, product_tmpl_id FROM exist e INNER JOIN product_product p ON p.id = e.product_id) INNER JOIN product_template pt ON product_tmpl_id = pt.id%s;",where)
	rows, err := r.DB.Query(query, offset, limit)
	if err != nil{
		fmt.Printf("Error en la consulta: %v\n", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next(){
		var product model.ProductDTO
		var stock sql.NullFloat64
		err := rows.Scan(&product.ID, &product.Name, &product.Price, &stock)
		if err != nil{
			fmt.Printf("Error to read row elemnt: %v\n",err)
			continue
		}
		product.Stock = stock.Float64
		Products = append(Products,product)
	}
	return Products, nil;
}

// GetBestSelling ordena por “sale_count” (campo de product.template) descendente y devuelve las variantes más vendidas.
// Para conseguir sale_count hay que leer primero del template.
func (r *odooProductRepo) GetBestSelling(limit int) ([]model.ProductDTO, error) {
	return []model.ProductDTO{}, nil;
}

// GetVariants busca todas las variantes del mismo template al que pertenece productID.
func (r *odooProductRepo) GetVariants(productID int64) ([]model.ProductDTO, error) {
	return []model.ProductDTO{}, nil;
}
