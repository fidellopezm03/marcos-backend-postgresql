package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/service"
)

// ProductHandler expone los endpoints HTTP relacionados con productos.
type ProductHandler struct {
	svc service.ProductService
}

// NewProductHandler inicializa el handler con el servicio.
func NewProductHandler(s service.ProductService) *ProductHandler {
	return &ProductHandler{svc: s}
}

// RegisterRoutes monta todas las rutas en el router pasado.
func (h *ProductHandler) RegisterRoutes(r chi.Router) {

	r.Route("/products", func(r chi.Router) {
		r.Get("/", h.getAll)                     // GET /products?page=&page_size=
		r.Get("/{id}", h.getByID)                // GET /products/{id}
		r.Post("/filtered", h.getFiltered)       // GET /products/filtered?categ_id=&min_price=&max_price=&page=&page_size=
		r.Post("/related", h.getRelated)         // GET /products/related?limit=
		r.Get("/best-selling", h.getBestSelling) // GET /products/best-selling?limit=
		r.Get("/{id}/variants", h.getVariants)   // GET /products/{id}/variants
		r.Get("/categories", h.getCategorys)
	})

}

// --- GET /products?page=&page_size= ---
func (h *ProductHandler) getAll(w http.ResponseWriter, r *http.Request) {
	// Leer query params (page, page_size)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	products, err := h.svc.GetAll(page, pageSize)
	if err != nil {
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}
	// Serializa Producto a JSON puro (se pueden mapear campos si se desea)
	render.JSON(w, r, products)
}

type Categories struct {
	Categories []string `json:"categories"`
}

// --- GET /products/filtered?categ_id=&min_price=&max_price=&page=&page_size= ---
func (h *ProductHandler) getFiltered(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Parsear categ_id si existe
	var categID *int64
	if raw := q.Get("categ_id"); raw != "" {
		if id64, err := strconv.ParseInt(raw, 10, 64); err == nil {
			categID = &id64
		}
	}

	// Parse price
	var minPrice, maxPrice *int64

	if raw := q.Get("min_price"); raw != "" {
		if i, err := strconv.ParseInt(raw, 10, 64); err == nil {
			minPrice = &i
		}
	}

	if raw := q.Get("max_price"); raw != "" {
		if i, err := strconv.ParseInt(raw, 10, 64); err == nil {
			maxPrice = &i
		}
	}

	orderValue := q.Get("order_value")

	// Paginación
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var categories Categories
	err := json.NewDecoder(r.Body).Decode(&categories)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Error in body request"})
		return
	}

	name := r.URL.Query().Get("name")
	products, err := h.svc.GetFiltered(page, pageSize, categID, minPrice, maxPrice, categories.Categories, name, orderValue)
	if err != nil {
		log.Printf("Error: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Error in server"})
		return
	}

	render.JSON(w, r, products)
}

// --- GET /products/{id}/related?limit= ---
func (h *ProductHandler) getRelated(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()
	page_size, err := strconv.Atoi(params.Get("page_size"))
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error en la solicitud"})
	}
	if page_size < 1 {
		page_size = 5
	}
	page, err := strconv.Atoi(params.Get("page"))
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error en la solicitud"})
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * page_size

	type bodyQuery struct {
		Name     string `json:"name"`
		Category string `json:"category"`
	}
	var body bodyQuery
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error en la solicitud"})
		return
	}
	products, err := h.svc.GetRelated(body.Category, body.Name, offset, page_size)
	if err != nil {
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}
	render.JSON(w, r, products)
}

func (h *ProductHandler) getByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	prodID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || prodID < 1 {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "id inválido"})
		return
	}

	product, err := h.svc.GetByID(prodID)
	if err != nil {
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}
	render.JSON(w, r, product)
}

// --- GET /products/best-selling?limit= ---
func (h *ProductHandler) getBestSelling(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	products, err := h.svc.GetBestSelling(limit)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Ha ocurrido un error en el servidor"})
		log.Printf("error: %v", err)
		return
	}
	render.JSON(w, r, products)
}
func (h *ProductHandler) getCategorys(w http.ResponseWriter, r *http.Request) {
	categorys, err := h.svc.GetCategorys()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"Error: ": err.Error()})
		return
	}
	render.JSON(w, r, categorys)
}

// --- GET /products/{id}/variants ---
func (h *ProductHandler) getVariants(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	prodID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || prodID < 1 {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "id inválido"})
		return
	}

	variants, err := h.svc.GetVariants(prodID)
	if err != nil {
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}
	render.JSON(w, r, variants)
}
