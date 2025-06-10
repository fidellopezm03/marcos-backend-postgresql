package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/model"
)
type Category struct{
	Category string
	CategoryName string
}
// ProductRepo define la interfaz para acceso a productos en Odoo.
type ProductRepo interface {
	GetAll(offset, limit int) (*model.ProductsResult, error)
	GetFiltered(offset, limit int, categID, minPrice, maxPrice *int64, categorys []string, name string) (*model.ProductsResult, error)
	GetRelated(productID int64, offset,limit *int) (*model.ProductsResult, error)
	GetBestSelling(limit int) ([]model.ProductDTO, error)
	GetVariants(productID int64) ([]model.ProductDTO, error)
	GetCategorys()([]Category, error)
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
func (r *odooProductRepo) GetAll(offset, limit int) (*model.ProductsResult, error) {
	return r.GetFiltered(offset,limit, nil, nil, nil, nil,"")
}

// GetRelated busca productos relacionados al productID dado.
func(r*odooProductRepo)GetRelated(productID int64, offset,limit *int) (*model.ProductsResult, error) {
	
	return &model.ProductsResult{}, nil;
}


func getValueJson(json, fallback string) (string, error){
	if len(json)<=0{ 
		return fallback, errors.New("error en la cadena entrante")
	}
	str := json[1:len(json)-1]
	lang:= strings.Split(str,",")
	if(len(lang)==2){
		str = lang[1]
	}else{
		str = lang[0]
	}
	str = strings.Split(str,":")[1]
	if str[0] == ' '{
		str = str[1:]
	}
	str = str[1:len(str)-1]	
	
	path := strings.Split(str,"/")
	str = path[len(path)-1]
	
	if str[0] == ' '{
		str = str[1:]
	}
	if str[len(str)-1] == ' '{
		str = str[:len(str)-1]
	}
	//fmt.Println(str)
	return str, nil
}
// GetFiltered permite filtrar por categoría (categ_id) y rango de precio list_price.
// Ambos filtros son opcionales: pasar nil para omitir.
func (r *odooProductRepo) GetFiltered(offset, limit int, categID, minPrice, maxPrice *int64, categorys []string, name string) (*model.ProductsResult, error) {
	var (
		params []any
	)
	ProductsResult := &model.ProductsResult{}
	i := 1
	//const id_location = "8"
	//selectQuery := []string{", SUM(quantity) as stock",", product.name as name, pc.name as category, list_price as price, stock", ", name, list_price, stock", ", stock"}
	exist := "WITH exist AS (SELECT product_id%s FROM stock_quant WHERE location_id = 8 GROUP BY product_id HAVING SUM(quantity) > 0"
	 
	where := ""
	queryp := "SELECT product_id as id%s FROM (SELECT product_id, pos_categ_id%s FROM (SELECT product_id%s, product_tmpl_id FROM exist e INNER JOIN product_product p ON p.id = e.product_id) INNER JOIN product_template pt ON product_tmpl_id = pt.id) product INNER JOIN pos_category pc ON pc.id = product.pos_categ_id%s"
	if minPrice != nil && maxPrice != nil {
		limitPrice := fmt.Sprintf("list_price >= %v and list_price <= %v",minPrice,maxPrice)
		where = " " + limitPrice + " "
	}
	if categorys != nil{
		if len(where)>0{
			where += " AND "
		}
		queryCateg := fmt.Sprintf(" pc.name->>'es_ES' = '%v'",i)
		i++
		for range len(categorys)-1{
			queryCateg+=fmt.Sprintf(" OR pc.name->>'es_ES' = '%v'",i)
			i++
		}
		if len(categorys)>1{
			queryCateg = "(" + queryCateg + " )"
		}

		where += queryCateg
		params = append(params, categorys)
		i++ 
	}
	if len(name)>0{
		if len(where)>0{
			where += " AND "
		}
		where += fmt.Sprintf("product.name LIKE '%$%v%'", i)
		params = append(params, name)
		i++
	}
	pag := fmt.Sprintf(" OFFSET %v LIMIT %v", offset, limit)
	
	whereCount := ""
	querypCount := queryp
	existCount := exist + ")"
	if len(where)>0{
		where =  " WHERE" + where
		whereCount = where
		queryp += pag
	}else{
		exist += pag
	}

	exist += ")"

	querypCount = "SELECT COUNT(*) as total FROM (" + fmt.Sprintf(querypCount,whereCount) + ");"
	queryCount := existCount + " " + queryp

	query := exist + " " + fmt.Sprintf(queryp,where)
	query+=";"

	errChan := make(chan error)
	
	go func(){
		rows, err := r.DB.Query(query,params...)
		if err != nil{
			log.Printf("Error en la consulta: %v\n", err)
			errChan <- err
			return 
		}
		defer rows.Close()
		for rows.Next(){
			var (
				product model.ProductDTO
				stock sql.NullFloat64
				name string
			
			)
		
			err := rows.Scan(&product.ID, &name, &product.CategoryName,&product.Price, &stock)
			if err != nil{
				log.Printf("Error to read row elemnt: %v\n",err)
				continue
			}
		
			if product.Name, err = getValueJson(name, "");err != nil {
				log.Printf("Error en name: %v",err)
			}
			if product.Category, err = getValueJson(product.CategoryName," "); err!=nil{
				log.Printf("Error en categ: %v",err)
			}
			product.Stock = stock.Float64
			ProductsResult.Products = append(ProductsResult.Products,product)
		}
		errChan<-nil
	}()
	
	row := r.DB.QueryRow(queryCount,params...)
	
	if(row == nil){
		err := fmt.Errorf("Error geting total products: %v",row.Err().Error())
		return nil,err
	}
	if err := row.Scan(&ProductsResult.Total);err!=nil{
		err := fmt.Errorf("Error scaning total product: %v", err)
		return nil, err
	}

	if err := <-errChan; err != nil{
		return nil,err
	}
	return ProductsResult, nil;
}
func (r *odooProductRepo) GetCategorys()([]Category,error){
	var categorys []Category
			
	row, err := r.DB.Query("SELECT name FROM pos_category")
	if err != nil{
		err = errors.New("error al obtenr las categorías")
		return nil, err
	}
	defer row.Close()
	for row.Next(){
		var category Category
		
		if err = row.Scan(&category.CategoryName);err!=nil{
			log.Printf("Error al leer el valor de category: %v",err)
			continue
		}
		if category.Category,err = getValueJson(category.CategoryName,"Any");err!=nil{
			log.Printf("error: %v", err)
		}
		categorys = append(categorys, category)
	}

	return categorys,nil
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
