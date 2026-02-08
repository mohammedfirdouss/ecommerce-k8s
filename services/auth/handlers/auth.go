package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/ecommerce/auth/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB        *sqlx.DB
	JWTSecret []byte
}

func NewAuthHandler(db *sqlx.DB) *AuthHandler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-me"
	}
	return &AuthHandler{DB: db, JWTSecret: []byte(secret)}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}

	var user models.User
	err = h.DB.QueryRowx(
		`INSERT INTO auth.users (email, password_hash) VALUES ($1, $2) 
		 RETURNING id, email, role, created_at`,
		req.Email, string(hash),
	).StructScan(&user)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already exists"})
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	writeJSON(w, http.StatusCreated, models.AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var user models.User
	err := h.DB.QueryRowx(
		`SELECT id, email, password_hash, role, created_at FROM auth.users WHERE email = $1`,
		req.Email,
	).StructScan(&user)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	writeJSON(w, http.StatusOK, models.AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) generateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.JWTSecret)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
