package models

import "time"

type Payment struct {
	ID        string    `json:"id" db:"id"`
	OrderID   string    `json:"order_id" db:"order_id"`
	Amount    float64   `json:"amount" db:"amount"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type OrderEvent struct {
	OrderID string  `json:"order_id"`
	UserID  string  `json:"user_id"`
	Total   float64 `json:"total"`
	Status  string  `json:"status"`
}

type PaymentStatusEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}
