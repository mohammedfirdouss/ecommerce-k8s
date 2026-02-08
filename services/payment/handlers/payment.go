package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ecommerce/payment/models"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentHandler struct {
	DB          *sqlx.DB
	Channel     *amqp.Channel
	SuccessRate float64
}

func NewPaymentHandler(db *sqlx.DB, ch *amqp.Channel) *PaymentHandler {
	rate := 0.8 // 80% success rate by default
	if v := os.Getenv("PAYMENT_SUCCESS_RATE"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			rate = parsed
		}
	}
	return &PaymentHandler{DB: db, Channel: ch, SuccessRate: rate}
}

func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	var payment models.Payment
	err := h.DB.QueryRowx(
		`SELECT id, order_id, amount, status, created_at, updated_at FROM payment.payments WHERE order_id = $1`,
		orderID,
	).StructScan(&payment)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "payment not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch payment"})
		return
	}
	writeJSON(w, http.StatusOK, payment)
}

// StartOrderConsumer listens for new order events and processes payments
func (h *PaymentHandler) StartOrderConsumer() {
	q, err := h.Channel.QueueDeclare("order.payments", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to declare queue: %v", err)
		return
	}

	err = h.Channel.QueueBind(q.Name, "order.created", "orders", false, nil)
	if err != nil {
		log.Printf("Failed to bind queue: %v", err)
		return
	}

	msgs, err := h.Channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to start consuming: %v", err)
		return
	}

	log.Println("Started order consumer for payment processing")
	for msg := range msgs {
		var event models.OrderEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("Failed to unmarshal order event: %v", err)
			continue
		}
		h.processPayment(event)
	}
}

func (h *PaymentHandler) processPayment(event models.OrderEvent) {
	log.Printf("Processing payment for order %s, amount: %.2f", event.OrderID, event.Total)

	// Simulate payment processing delay
	time.Sleep(2 * time.Second)

	// Mock payment: random success/failure based on success rate
	status := "completed"
	if rand.Float64() > h.SuccessRate {
		status = "failed"
	}

	// Store payment record
	_, err := h.DB.Exec(
		`INSERT INTO payment.payments (order_id, amount, status) VALUES ($1, $2, $3)
		 ON CONFLICT (order_id) DO UPDATE SET status = $3, updated_at = NOW()`,
		event.OrderID, event.Total, status,
	)
	if err != nil {
		log.Printf("Failed to store payment: %v", err)
		// Still publish status update
	}

	// Publish payment status back to order service
	statusEvent := models.PaymentStatusEvent{
		OrderID: event.OrderID,
		Status:  status,
	}
	body, err := json.Marshal(statusEvent)
	if err != nil {
		log.Printf("Failed to marshal status event: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.Channel.PublishWithContext(ctx,
		"payments",
		"payment.status",
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("Failed to publish payment status: %v", err)
	} else {
		log.Printf("Payment for order %s: %s", event.OrderID, status)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
