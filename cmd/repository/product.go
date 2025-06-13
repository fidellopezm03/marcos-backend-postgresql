package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/model"
)
type Category struct{
	Category string `json:"category"`
	CategoryName string `json:"categoryName"`
}
// ProductRepo define la interfaz para acceso a productos en Odoo.
type ProductRepo interface {
	GetAll(offset, limit int) (*model.ProductsResult, error)
	GetFiltered(offset, limit int, categID, minPrice, maxPrice *int64, categorys []string, name, orderValue string) (*model.ProductsResult, error)
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
	return r.GetFiltered(offset,limit, nil, nil, nil, nil,"","")
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
func (r *odooProductRepo) GetFiltered(offset, limit int, categID, minPrice, maxPrice *int64, categorys []string, name, orderValue string) (*model.ProductsResult, error) {
	var params, anys []any
	stringsToanys := func (strings []string)[]any{
		anys := make([]any,len(strings))
		for index, value := range strings{
			anys[index]=value
		}
		return anys
	}
	wheres := []string{}
	ProductsResult := &model.ProductsResult{}
	i := 1
	//const id_location = "8"
	selectQuery := []string{", SUM(quantity) as stock",", product.name as name, pc.name as category, list_price as price, stock", ", name, list_price, stock", ", stock"}
	selectQueryCount := []string{}
	
	for range selectQuery{
		selectQueryCount = append(selectQueryCount, "")
	}

	exist := "WITH exist AS (SELECT product_id%s FROM stock_quant WHERE location_id = 8 GROUP BY product_id HAVING SUM(quantity) > 0"
	 
	const where = " WHERE"
	
	queryp := "SELECT product_id as id%s FROM (SELECT product_id, categ_id%s FROM (SELECT product_id%s, product_tmpl_id FROM exist e INNER JOIN product_product p ON p.id = e.product_id) INNER JOIN product_template pt ON product_tmpl_id = pt.id%s) product LEFT JOIN product_category pc ON pc.id = product.categ_id%s"
	queryWhere := ""

	if minPrice != nil && maxPrice != nil {
		limitPrice := fmt.Sprintf(" list_price >= %v and list_price <= %v",minPrice,maxPrice)
		queryWhere += limitPrice
		
	}
	if len(name)>0{
		if len(queryWhere) > 0{
			queryWhere += " AND "
		}
		queryWhere += fmt.Sprintf("product.name LIKE '%%' || $%d || '%%'", i)
		queryWhere = where + queryWhere
		params = append(params, name)
		i++
	}

	orderBy := ""

	if strings.ToLower(orderValue) == "desc" || strings.ToLower(orderValue) == "asc" {
		orderBy = fmt.Sprintf(" GROUP BY list_price %s", strings.ToUpper(orderValue))
	}
	wheres = append(wheres, queryWhere + orderBy)
	
	queryWhere = ""
	if categorys != nil{
		
		queryWhere = fmt.Sprintf(" pc.name LIKE '%%'|| $%d || '%%'",i)
		i++
		for range len(categorys)-1{
			queryWhere+=fmt.Sprintf(" OR pc.name LIKE '%%' || $%d || '%%'",i)
			i++
		}
		if len(categorys)>1{
			queryWhere = where + " (" + queryWhere + " )"
		}
		
		anys := stringsToanys(categorys)
		params = append(params, anys...)
	}
	wheres = append(wheres, queryWhere)
	
	pag := fmt.Sprintf(" OFFSET %v LIMIT %v", offset, limit)
	
	
	querypCount := queryp

	existCount := exist + ")"
	
	queryCount := ""

	if len(params)==0{
		queryCount = fmt.Sprintf(existCount + " SELECT COUNT(*) as total FROM exist;","")
	}else{
		anys = stringsToanys(append(selectQueryCount,wheres...))
		queryCount = fmt.Sprintf(existCount + " SELECT COUNT(*) as total FROM (" + querypCount + ");",anys...)
	}
	if len(where)>0{
		queryp += pag
	}else{
		exist += pag
	}

	exist += ")"
	
	anys = stringsToanys(append(selectQuery,wheres...))
	query := fmt.Sprintf(exist + " " + queryp,anys...)
	query+=";"

	var (
		wg sync.WaitGroup
		errQueryProducts error
	)
	wg.Add(1)
	go func(){
		defer wg.Done()
		rows, err := r.DB.Query(query,params...)
		if err != nil{
			errQueryProducts = fmt.Errorf("Error en la consulta: %v\n", err)
			return 
		}
		defer rows.Close()
		for rows.Next(){
			var (
				product model.ProductDTO
				stock sql.NullFloat64
				category sql.NullString
				name string
			)
		
			err := rows.Scan(&product.ID, &name, &category,&product.OriginalPrice, &stock)
			if err != nil{
				log.Printf("Error to read row elemnt: %v\n",err)
				continue
			}
			product.Price = product.OriginalPrice
		
			if product.Name, err = getValueJson(name, "");err != nil {
				log.Printf("Error en name: %v",err)
			}
			if category.Valid{
				product.CategoryName = category.String
				strs := strings.Split(product.CategoryName, "/")
				product.Category = strs[len(strs)-1]
			}
			product.Stock = stock.Float64
			ProductsResult.Products = append(ProductsResult.Products,product)
		}
		errQueryProducts = nil
	}()
	
	row := r.DB.QueryRow(queryCount,params...)
	
	if(row == nil){
		err := fmt.Errorf("Error geting total products: %v",row.Err())
		return nil,err
	}
	// if err := row.Scan(&ProductsResult.Total);err!=nil{
	// 	err = fmt.Errorf("Error scaning total product: %v", err)
	// 	return nil, err
	// }
	wg.Wait()
	if errQueryProducts != nil{
		return nil,errQueryProducts
	}
	return ProductsResult, nil;
}
func (r *odooProductRepo) GetCategorys()([]Category,error){
	var categorys []Category
			
	row, err := r.DB.Query("SELECT name FROM product_category")
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
		strs := strings.Split(category.CategoryName,"/")
		category.Category = strs[len(strs)-1]
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
