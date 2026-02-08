package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ecommerce/order/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	dbHost := getEnv("DB_HOST", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "ecommerce")
	dbPass := getEnv("DB_PASSWORD", "ecommerce_pass")
	dbName := getEnv("DB_NAME", "ecommerce")
	port := getEnv("PORT", "8084")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	// Declare exchanges
	ch.ExchangeDeclare("orders", "topic", true, false, false, false, nil)
	ch.ExchangeDeclare("payments", "topic", true, false, false, false, nil)

	orderHandler := handlers.NewOrderHandler(db, ch)

	// Start background consumer for payment status updates
	go orderHandler.StartPaymentStatusConsumer()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	r.Handle("/metrics", promhttp.Handler())

	r.Post("/", orderHandler.CreateOrder)
	r.Get("/", orderHandler.ListOrders)
	r.Get("/{id}", orderHandler.GetOrder)

	log.Printf("Order service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
