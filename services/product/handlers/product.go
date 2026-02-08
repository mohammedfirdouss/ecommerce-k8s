package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/ecommerce/product/models"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type ProductHandler struct {
	DB *sqlx.DB
}

func NewProductHandler(db *sqlx.DB) *ProductHandler {
	return &ProductHandler{DB: db}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	var products []models.Product
	err := h.DB.Select(&products, `SELECT id, name, description, price, stock, created_at, updated_at FROM product.products ORDER BY created_at DESC`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch products"})
		return
	}
	if products == nil {
		products = []models.Product{}
	}
	writeJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var product models.Product
	err := h.DB.QueryRowx(
		`SELECT id, name, description, price, stock, created_at, updated_at FROM product.products WHERE id = $1`, id,
	).StructScan(&product)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch product"})
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Name == "" || req.Price <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and positive price are required"})
		return
	}

	var product models.Product
	err := h.DB.QueryRowx(
		`INSERT INTO product.products (name, description, price, stock) VALUES ($1, $2, $3, $4) 
		 RETURNING id, name, description, price, stock, created_at, updated_at`,
		req.Name, req.Description, req.Price, req.Stock,
	).StructScan(&product)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create product"})
		return
	}
	writeJSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req models.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var product models.Product
	err := h.DB.QueryRowx(
		`UPDATE product.products SET 
			name = COALESCE($1, name),
			description = COALESCE($2, description),
			price = COALESCE($3, price),
			stock = COALESCE($4, stock),
			updated_at = NOW()
		 WHERE id = $5
		 RETURNING id, name, description, price, stock, created_at, updated_at`,
		req.Name, req.Description, req.Price, req.Stock, id,
	).StructScan(&product)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update product"})
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := h.DB.Exec(`DELETE FROM product.products WHERE id = $1`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete product"})
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "product deleted"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
