package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ecommerce/cart/models"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type CartHandler struct {
	DB *sqlx.DB
}

func NewCartHandler(db *sqlx.DB) *CartHandler {
	return &CartHandler{DB: db}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing user ID"})
		return
	}

	// Get or create cart
	var cart models.Cart
	err := h.DB.QueryRowx(
		`SELECT id, user_id, created_at, updated_at FROM cart.carts WHERE user_id = $1`, userID,
	).StructScan(&cart)
	if err != nil {
		// Create new cart
		err = h.DB.QueryRowx(
			`INSERT INTO cart.carts (user_id) VALUES ($1) RETURNING id, user_id, created_at, updated_at`,
			userID,
		).StructScan(&cart)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create cart"})
			return
		}
	}

	// Get cart items
	var items []models.CartItem
	err = h.DB.Select(&items,
		`SELECT id, cart_id, product_id, quantity, price, created_at FROM cart.cart_items WHERE cart_id = $1`,
		cart.ID,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch cart items"})
		return
	}
	if items == nil {
		items = []models.CartItem{}
	}
	cart.Items = items

	writeJSON(w, http.StatusOK, cart)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing user ID"})
		return
	}

	var req models.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.ProductID == "" || req.Quantity <= 0 || req.Price <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "product_id, positive quantity and price are required"})
		return
	}

	// Get or create cart
	var cartID string
	err := h.DB.QueryRow(`SELECT id FROM cart.carts WHERE user_id = $1`, userID).Scan(&cartID)
	if err != nil {
		err = h.DB.QueryRow(
			`INSERT INTO cart.carts (user_id) VALUES ($1) RETURNING id`, userID,
		).Scan(&cartID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create cart"})
			return
		}
	}

	// Check if item already in cart, update quantity
	var existingID string
	err = h.DB.QueryRow(
		`SELECT id FROM cart.cart_items WHERE cart_id = $1 AND product_id = $2`, cartID, req.ProductID,
	).Scan(&existingID)
	if err == nil {
		// Update quantity
		var item models.CartItem
		err = h.DB.QueryRowx(
			`UPDATE cart.cart_items SET quantity = quantity + $1 WHERE id = $2 
			 RETURNING id, cart_id, product_id, quantity, price, created_at`,
			req.Quantity, existingID,
		).StructScan(&item)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update item"})
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}

	// Insert new item
	var item models.CartItem
	err = h.DB.QueryRowx(
		`INSERT INTO cart.cart_items (cart_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)
		 RETURNING id, cart_id, product_id, quantity, price, created_at`,
		cartID, req.ProductID, req.Quantity, req.Price,
	).StructScan(&item)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add item"})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")
	result, err := h.DB.Exec(`DELETE FROM cart.cart_items WHERE id = $1`, itemID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove item"})
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "item removed"})
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing user ID"})
		return
	}

	_, err := h.DB.Exec(
		`DELETE FROM cart.cart_items WHERE cart_id IN (SELECT id FROM cart.carts WHERE user_id = $1)`,
		userID,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to clear cart"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "cart cleared"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
