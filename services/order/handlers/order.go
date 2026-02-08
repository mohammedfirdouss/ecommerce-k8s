package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ecommerce/order/models"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderHandler struct {
	DB      *sqlx.DB
	Channel *amqp.Channel
}

func NewOrderHandler(db *sqlx.DB, ch *amqp.Channel) *OrderHandler {
	return &OrderHandler{DB: db, Channel: ch}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing user ID"})
		return
	}

	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if len(req.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "at least one item is required"})
		return
	}

	// Calculate total
	var total float64
	for _, item := range req.Items {
		total += item.Price * float64(item.Quantity)
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Create order
	var order models.Order
	err = tx.QueryRowx(
		`INSERT INTO orders.orders (user_id, total) VALUES ($1, $2) 
		 RETURNING id, user_id, total, status, created_at, updated_at`,
		userID, total,
	).StructScan(&order)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create order"})
		return
	}

	// Create order items
	for _, item := range req.Items {
		_, err = tx.Exec(
			`INSERT INTO orders.order_items (order_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)`,
			order.ID, item.ProductID, item.Quantity, item.Price,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create order item"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to commit transaction"})
		return
	}

	// Publish order event to RabbitMQ
	event := models.OrderEvent{
		OrderID: order.ID,
		UserID:  userID,
		Total:   total,
		Status:  "pending",
	}
	h.publishOrderEvent(event)

	// Fetch items for response
	var items []models.OrderItem
	h.DB.Select(&items, `SELECT id, order_id, product_id, quantity, price FROM orders.order_items WHERE order_id = $1`, order.ID)
	if items == nil {
		items = []models.OrderItem{}
	}
	order.Items = items

	writeJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing user ID"})
		return
	}

	var orders []models.Order
	err := h.DB.Select(&orders,
		`SELECT id, user_id, total, status, created_at, updated_at FROM orders.orders WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch orders"})
		return
	}
	if orders == nil {
		orders = []models.Order{}
	}
	writeJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var order models.Order
	err := h.DB.QueryRowx(
		`SELECT id, user_id, total, status, created_at, updated_at FROM orders.orders WHERE id = $1`, id,
	).StructScan(&order)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch order"})
		return
	}

	var items []models.OrderItem
	h.DB.Select(&items, `SELECT id, order_id, product_id, quantity, price FROM orders.order_items WHERE order_id = $1`, order.ID)
	if items == nil {
		items = []models.OrderItem{}
	}
	order.Items = items

	writeJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) publishOrderEvent(event models.OrderEvent) {
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal order event: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.Channel.PublishWithContext(ctx,
		"orders",        // exchange
		"order.created", // routing key
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("Failed to publish order event: %v", err)
	}
}

// StartPaymentStatusConsumer listens for payment status updates from RabbitMQ
func (h *OrderHandler) StartPaymentStatusConsumer() {
	q, err := h.Channel.QueueDeclare("payment.status", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to declare queue: %v", err)
		return
	}

	err = h.Channel.QueueBind(q.Name, "payment.status", "payments", false, nil)
	if err != nil {
		log.Printf("Failed to bind queue: %v", err)
		return
	}

	msgs, err := h.Channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to start consuming: %v", err)
		return
	}

	log.Println("Started payment status consumer")
	for msg := range msgs {
		var event models.OrderEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("Failed to unmarshal payment event: %v", err)
			continue
		}
		_, err := h.DB.Exec(
			`UPDATE orders.orders SET status = $1, updated_at = NOW() WHERE id = $2`,
			event.Status, event.OrderID,
		)
		if err != nil {
			log.Printf("Failed to update order status: %v", err)
		} else {
			log.Printf("Updated order %s status to %s", event.OrderID, event.Status)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
