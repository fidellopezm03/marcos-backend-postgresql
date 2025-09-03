package repository

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/model"
)

type Category struct {
	Category     string `json:"category"`
	CategoryName string `json:"categoryName"`
}

// ProductRepo define la interfaz para acceso a productos en Odoo.
type ProductRepo interface {
	GetAll(offset, limit int) (*model.ProductsResult, error)
	GetByID(id int64) (*model.ProductDTO, error)
	GetFiltered(offset, limit int, categID, minPrice, maxPrice *int64, categorys []string, name, orderValue string) (*model.ProductsResult, error)
	GetRelated(category, name string, offset, limit *int) (*model.ProductsResult, error)
	GetBestSelling(limit int) ([]model.ProductDTO, error)
	GetVariants(productID int64) ([]model.ProductDTO, error)
	GetCategorys() ([]Category, error)
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
	return r.GetFiltered(offset, limit, nil, nil, nil, nil, "", "")
}
func (r *odooProductRepo) GetByID(id int64) (*model.ProductDTO, error) {
	query := fmt.Sprintf("WITH exist AS (SELECT product_id, SUM(quantity) as stock FROM stock_quant WHERE location_id = 8 AND product_id = %d  GROUP BY product_id HAVING SUM(quantity) > 0) SELECT product_id as id, product.name as name, pc.name as category, list_price as price, stock FROM (SELECT product_id, categ_id, name, list_price, stock FROM (SELECT product_id, stock, product_tmpl_id FROM exist e INNER JOIN product_product p ON p.id = e.product_id) INNER JOIN product_template pt ON product_tmpl_id = pt.id) product LEFT JOIN product_category pc ON pc.id = product.categ_id;", id)
	row := r.DB.QueryRow(query)
	if row == nil {
		return nil, errors.New("error al obtener producto")
	}

	var (
		product  model.ProductDTO
		stock    sql.NullFloat64
		category sql.NullString
		name     string
	)

	err := row.Scan(&product.ID, &name, &category, &product.OriginalPrice, &stock)
	if err != nil {
		log.Printf("Error to read row elemnt: %v\n", err)
	}
	product.Price = product.OriginalPrice

	if product.Name, err = getValueJson(name, ""); err != nil {
		log.Printf("Error en name: %v", err)
	}
	if category.Valid {
		product.CategoryName = category.String
		strs := strings.Split(product.CategoryName, "/")
		product.Category = strs[len(strs)-1]
		if product.Category[0] == ' ' {
			product.Category = product.Category[1:]
		}
	}
	product.Stock = stock.Float64
	return &product, nil

}

// GetRelated busca productos relacionados al productID dado.
func (r *odooProductRepo) GetRelated(category, name string, offset, limit *int) (*model.ProductsResult, error) {
	return r.GetFiltered(*offset, *limit, nil, nil, nil, []string{category}, name, "")
}

func getValueJson(json, fallback string) (string, error) {
	if len(json) <= 0 {
		return fallback, errors.New("error en la cadena entrante")
	}
	str := json[1 : len(json)-1]
	lang := strings.Split(str, ",")
	if len(lang) == 2 {
		str = lang[1]
	} else {
		str = lang[0]
	}
	str = strings.Split(str, ":")[1]
	if str[0] == ' ' {
		str = str[1:]
	}
	str = str[1 : len(str)-1]

	path := strings.Split(str, "/")
	str = path[len(path)-1]

	if str[0] == ' ' {
		str = str[1:]
	}
	if str[len(str)-1] == ' ' {
		str = str[:len(str)-1]
	}
	//fmt.Println(str)
	return str, nil
}

// GetFiltered permite filtrar por categoría (categ_id) y rango de precio list_price.
// Ambos filtros son opcionales: pasar nil para omitir.
func (r *odooProductRepo) GetFiltered(offset, limit int, categID, minPrice, maxPrice *int64, categorys []string, name, orderValue string) (*model.ProductsResult, error) {
	var params, anys []any
	stringsToanys := func(strings []string) []any {
		anys := make([]any, len(strings))
		for index, value := range strings {
			anys[index] = value
		}
		return anys
	}
	wheres := []string{}
	ProductsResult := &model.ProductsResult{}
	i := 1
	//const id_location = "8"
	selectQuery := []string{", SUM(quantity) as stock", ", product.name as name, pc.name as category, list_price as price, stock", ", name, list_price, stock", ", stock"}
	selectQueryCount := []string{}

	for range selectQuery {
		selectQueryCount = append(selectQueryCount, "")
	}

	exist := "WITH exist AS (SELECT product_id%s FROM stock_quant WHERE location_id = 8 GROUP BY product_id HAVING SUM(quantity) > 0"

	const where = " WHERE"

	queryp := "SELECT product_id as id%s FROM (SELECT product_id, categ_id%s FROM (SELECT product_id%s, product_tmpl_id FROM exist e INNER JOIN product_product p ON p.id = e.product_id) INNER JOIN product_template pt ON product_tmpl_id = pt.id%s) product LEFT JOIN product_category pc ON pc.id = product.categ_id%s"
	queryWhere := ""

	if minPrice != nil && maxPrice != nil {
		limitPrice := fmt.Sprintf(" list_price >= %v and list_price <= %v", minPrice, maxPrice)
		queryWhere += where + limitPrice

	}
	if len(name) > 0 {
		if len(queryWhere) > 0 {
			queryWhere += " AND"
		} else {
			queryWhere += where
		}
		queryWhere += fmt.Sprintf(" product.name LIKE '%%' || $%d || '%%'", i)
		params = append(params, name)
		i++
	}

	orderBy := ""

	if value := strings.ToLower(orderValue); value == "desc" || value == "asc" {
		orderBy = fmt.Sprintf(" ORDER BY list_price %s", value)
	}
	wheres = append(wheres, queryWhere+orderBy)

	queryWhere = ""
	if len(categorys) > 0 {

		queryWhere = fmt.Sprintf("pc.name LIKE '%%'|| $%d || '%%'", i)
		i++
		for range len(categorys) - 1 {
			queryWhere += fmt.Sprintf(" OR pc.name LIKE '%%' || $%d || '%%'", i)
			i++
		}
		queryWhere = where + " ( " + queryWhere + " )"

		anys := stringsToanys(categorys)
		params = append(params, anys...)
	}
	wheres = append(wheres, queryWhere)

	pag := fmt.Sprintf(" OFFSET %v LIMIT %v", offset, limit)

	querypCount := queryp

	existCount := exist + ")"

	queryCount := ""

	if len(wheres) == 0 {
		queryCount = fmt.Sprintf(existCount+" SELECT COUNT(*) as total FROM exist;", "")
	} else {
		anys = stringsToanys(append(selectQueryCount, wheres...))
		queryCount = fmt.Sprintf(existCount+" SELECT COUNT(*) as total FROM ("+querypCount+");", anys...)
	}
	if len(where) > 0 {
		queryp += pag
	} else {
		exist += pag
	}

	exist += ")"

	anys = stringsToanys(append(selectQuery, wheres...))
	query := fmt.Sprintf(exist+" "+queryp, anys...)
	query += ";"

	var (
		wg               sync.WaitGroup
		errQueryProducts error
		ProductsMap      = make(map[uint64]*model.ProductDTO)
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		rows, err := r.DB.Query(query, params...)
		if err != nil {
			errQueryProducts = fmt.Errorf("error en la consulta: %v", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var (
				stock    sql.NullFloat64
				category sql.NullString
				name     string
			)
			product := &model.ProductDTO{}

			err := rows.Scan(&product.ID, &name, &category, &product.OriginalPrice, &stock)
			if err != nil {
				log.Printf("Error to read row elemnt: %v\n", err)
				continue
			}
			product.Price = product.OriginalPrice

			if product.Name, err = getValueJson(name, ""); err != nil {
				log.Printf("Error en name: %v", err)
			}
			if category.Valid {
				product.CategoryName = category.String
				strs := strings.Split(product.CategoryName, "/")
				product.Category = strs[len(strs)-1]
				if product.Category[0] == ' ' {
					product.Category = product.Category[1:]
				}
			}
			product.Stock = stock.Float64
			ProductsMap[product.ID] = product
		}
		errQueryProducts = nil
	}()

	row := r.DB.QueryRow(queryCount, params...)

	if row == nil {
		err := fmt.Errorf("error geting total products: %v", row.Err())
		return nil, err
	}
	if err := row.Scan(&ProductsResult.Total); err != nil {
		err = fmt.Errorf("error scaning total product: %v", err)
		return nil, err
	}
	if ProductsResult.Total == 0 {
		return nil, errors.New("no products found")
	}
	wg.Wait()

	if errQueryProducts != nil {
		return nil, errQueryProducts
	}

	query = "SELECT res_id, mimetype, db_datas FROM ir_attachment WHERE db_dates NOT NULL AND length(db_datas) > 0 AND ("
	for id := range ProductsMap {
		query += fmt.Sprintf(" res_id = %d OR", id)
	}
	query = query[:len(query)-3] + " );"
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener las imágenes: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id       uint64
			mime     string
			db_datas []byte
		)
		if err = rows.Scan(&id, mime, &db_datas); err != nil {
			log.Printf("error al leer el valor de db_datas: %v", err)
			continue
		}
		if mime != "image/png" && mime != "image/jpeg" {
			continue
		}
		base64Str := base64.StdEncoding.EncodeToString(db_datas)
		ProductsMap[id].Images = append(ProductsMap[id].Images, fmt.Sprintf("data:%s;base64,", mime)+base64Str)
	}
	for _, product := range ProductsMap {
		ProductsResult.Products = append(ProductsResult.Products, *product)
	}

	return ProductsResult, nil
}
func (r *odooProductRepo) GetCategorys() ([]Category, error) {
	var categorys []Category

	row, err := r.DB.Query("SELECT name FROM product_category")
	if err != nil {
		err = errors.New("error al obtener las categorías")
		return nil, err
	}
	defer row.Close()
	exist := func(category string) bool {
		for _, cat := range categorys {
			if category == cat.Category {
				return true
			}
		}
		return false
	}
	for row.Next() {
		var category Category

		if err = row.Scan(&category.CategoryName); err != nil {
			log.Printf("error al leer el valor de category: %v", err)
			continue
		}
		strs := strings.Split(category.CategoryName, "/")
		if len(strs) == 2 || len(strs) == 3 {
			continue
		}
		category.Category = strs[len(strs)-1]
		if category.Category[0] == ' ' {
			category.Category = category.Category[1:]
		}
		if !exist(category.Category) {
			categorys = append(categorys, category)
		}

	}

	return categorys, nil
}

// GetBestSelling ordena por “sale_count” (campo de product.template) descendente y devuelve las variantes más vendidas.
// Para conseguir sale_count hay que leer primero del template.
func (r *odooProductRepo) GetBestSelling(limit int) ([]model.ProductDTO, error) {
	query := fmt.Sprintf("WITH exist AS (SELECT product_id, SUM(quantity) as stock FROM stock_quant WHERE location_id = 8 GROUP BY product_id having sum(quantity)>0) select product_id, pct.name as name , pc.name as category, price, stock  from (select product_id, stock, categ_id, name, list_price as price, quantity from (select product_id, stock, product_tmpl_id, quantity from (select exist.product_id, stock, quantity from (select product_id, sum(quantity_done) as quantity from stock_move where location_dest_id = 5 group by product_id) l inner join exist on l.product_id = exist.product_id) o inner join product_product pp on pp.id = o.product_id) ptl inner join product_template pt on pt.id = ptl.product_tmpl_id) pct inner join product_category pc on pct.categ_id = pc.id order by quantity desc limit %d", limit)

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ha ocurrido un error: %v", err)
	}
	defer rows.Close()
	var Products []model.ProductDTO
	for rows.Next() {
		var (
			product  model.ProductDTO
			stock    sql.NullFloat64
			category sql.NullString
			name     string
		)

		err := rows.Scan(&product.ID, &name, &category, &product.OriginalPrice, &stock)
		if err != nil {
			log.Printf("Error to read row elemnt: %v\n", err)
			continue
		}
		product.Price = product.OriginalPrice

		if product.Name, err = getValueJson(name, ""); err != nil {
			log.Printf("Error en name: %v", err)
		}
		if category.Valid {
			product.CategoryName = category.String
			strs := strings.Split(product.CategoryName, "/")
			product.Category = strs[len(strs)-1]
			if product.Category[0] == ' ' {
				product.Category = product.Category[1:]
			}
		}
		product.Stock = stock.Float64
		Products = append(Products, product)
	}
	return Products, nil
}

// GetVariants busca todas las variantes del mismo template al que pertenece productID.
func (r *odooProductRepo) GetVariants(productID int64) ([]model.ProductDTO, error) {
	return []model.ProductDTO{}, nil
}
