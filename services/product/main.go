package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ecommerce/product/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	dbHost := getEnv("DB_HOST", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "ecommerce")
	dbPass := getEnv("DB_PASSWORD", "ecommerce_pass")
	dbName := getEnv("DB_NAME", "ecommerce")
	port := getEnv("PORT", "8082")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	productHandler := handlers.NewProductHandler(db)

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

	r.Get("/", productHandler.List)
	r.Get("/{id}", productHandler.Get)
	r.Post("/", productHandler.Create)
	r.Put("/{id}", productHandler.Update)
	r.Delete("/{id}", productHandler.Delete)

	log.Printf("Product service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
