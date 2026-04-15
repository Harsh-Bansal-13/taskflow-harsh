package handler

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/models"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/repository"
)

type UserHandler struct {
	userRepo repository.UserRepo
	logger   *slog.Logger
}

func NewUserHandler(userRepo repository.UserRepo, logger *slog.Logger) *UserHandler {
	return &UserHandler{userRepo: userRepo, logger: logger}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	var err error

	projectIDStr := r.URL.Query().Get("project_id")
	if projectIDStr == "" {
		writeError(w, http.StatusBadRequest, "validation failed", map[string]string{"project_id": "is required"})
		return
	}

	projectID, parseErr := uuid.Parse(projectIDStr)
	if parseErr != nil {
		writeError(w, http.StatusBadRequest, "validation failed", map[string]string{"project_id": "invalid uuid"})
		return
	}
	users, err = h.userRepo.ListByProject(r.Context(), projectID)

	if err != nil {
		h.logger.Error("list users", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if users == nil {
		users = []models.User{}
	}
	writeJSON(w, http.StatusOK, users)
}
