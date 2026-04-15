package handler

import (
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/models"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type AuthHandler struct {
	userRepo  repository.UserRepo
	jwtSecret string
	logger    *slog.Logger
}

func NewAuthHandler(userRepo repository.UserRepo, jwtSecret string, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwtSecret: jwtSecret, logger: logger}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	// Validate
	fields := make(map[string]string)
	if strings.TrimSpace(req.Name) == "" {
		fields["name"] = "is required"
	}
	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "is required"
	} else if !emailRegex.MatchString(strings.TrimSpace(req.Email)) {
		fields["email"] = "invalid format"
	}
	if len(req.Password) < 8 {
		fields["password"] = "must be at least 8 characters"
	}
	if len(fields) > 0 {
		writeError(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	// Check existing
	existing, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		h.logger.Error("register: check existing", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if existing != nil {
		writeError(w, http.StatusBadRequest, "validation failed", map[string]string{"email": "already registered"})
		return
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		h.logger.Error("register: hash password", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	user := &models.User{
		ID:       uuid.New(),
		Name:     strings.TrimSpace(req.Name),
		Email:    strings.TrimSpace(req.Email),
		Password: string(hashed),
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		h.logger.Error("register: create user", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		h.logger.Error("register: generate token", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	writeJSON(w, http.StatusCreated, models.AuthResponse{Token: token, User: *user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	fields := make(map[string]string)
	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "is required"
	}
	if req.Password == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		writeError(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), strings.TrimSpace(req.Email))
	if err != nil {
		h.logger.Error("login: get user", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password", nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password", nil)
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		h.logger.Error("login: generate token", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	writeJSON(w, http.StatusOK, models.AuthResponse{Token: token, User: *user})
}

func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
