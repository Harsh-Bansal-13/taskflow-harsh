package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/broker"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/middleware"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/models"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/repository"
)

type ProjectHandler struct {
	projectRepo repository.ProjectRepo
	taskRepo    repository.TaskRepo
	broker      *broker.EventBroker
	logger      *slog.Logger
}

func NewProjectHandler(projectRepo repository.ProjectRepo, taskRepo repository.TaskRepo, b *broker.EventBroker, logger *slog.Logger) *ProjectHandler {
	return &ProjectHandler{projectRepo: projectRepo, taskRepo: taskRepo, broker: b, logger: logger}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	page, limit := parsePagination(r)

	projects, total, err := h.projectRepo.ListByUser(r.Context(), userID, page, limit)
	if err != nil {
		h.logger.Error("list projects", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if projects == nil {
		projects = []models.Project{}
	}

	writeJSON(w, http.StatusOK, paginatedResponse(projects, page, limit, total))
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "validation failed", map[string]string{"name": "is required"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	project := &models.Project{
		ID:          uuid.New(),
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		OwnerID:     userID,
	}

	if err := h.projectRepo.Create(r.Context(), project); err != nil {
		h.logger.Error("create project", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id", nil)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("get project", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if project == nil {
		writeError(w, http.StatusNotFound, "not found", nil)
		return
	}

	userID := middleware.GetUserID(r.Context())
	isMember, memberErr := h.projectRepo.IsMember(r.Context(), id, userID)
	if memberErr != nil {
		h.logger.Error("get project: check membership", "error", memberErr)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "forbidden", nil)
		return
	}

	// Fetch all tasks for this project (no row cap)
	tasks, err := h.taskRepo.GetAllByProject(r.Context(), id)
	if err != nil {
		h.logger.Error("get project tasks", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if tasks == nil {
		tasks = []models.Task{}
	}

	result := models.ProjectWithTasks{
		Project: *project,
		Tasks:   tasks,
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id", nil)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("update project: get", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if project == nil {
		writeError(w, http.StatusNotFound, "not found", nil)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if project.OwnerID != userID {
		writeError(w, http.StatusForbidden, "forbidden", nil)
		return
	}

	var req models.UpdateProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	updated, err := h.projectRepo.Update(r.Context(), id, req.Name, req.Description)
	if err != nil {
		h.logger.Error("update project", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id", nil)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("delete project: get", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if project == nil {
		writeError(w, http.StatusNotFound, "not found", nil)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if project.OwnerID != userID {
		writeError(w, http.StatusForbidden, "forbidden", nil)
		return
	}

	if err := h.projectRepo.Delete(r.Context(), id); err != nil {
		h.logger.Error("delete project", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) Stats(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id", nil)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("project stats: get", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if project == nil {
		writeError(w, http.StatusNotFound, "not found", nil)
		return
	}

	userID := middleware.GetUserID(r.Context())
	isMember, memberErr := h.projectRepo.IsMember(r.Context(), id, userID)
	if memberErr != nil {
		h.logger.Error("project stats: check membership", "error", memberErr)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "forbidden", nil)
		return
	}

	stats, err := h.taskRepo.GetProjectStats(r.Context(), id)
	if err != nil {
		h.logger.Error("project stats", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// Events streams Server-Sent Events for task mutations on a project.
// Auth is via ?token= query param since EventSource cannot set custom headers.
func (h *ProjectHandler) Events(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id", nil)
		return
	}

	// Verify the project exists and the caller owns it
	project, err := h.projectRepo.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("events: get project", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if project == nil {
		writeError(w, http.StatusNotFound, "not found", nil)
		return
	}
	userID := middleware.GetUserID(r.Context())
	isMember, memberErr := h.projectRepo.IsMember(r.Context(), id, userID)
	if memberErr != nil {
		h.logger.Error("events: check membership", "error", memberErr)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "forbidden", nil)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	// Send initial connected event
	fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	flusher.Flush()

	ch := h.broker.Subscribe(id)
	defer h.broker.Unsubscribe(id, ch)

	for {
		select {
		case <-ch:
			fmt.Fprintf(w, "event: task_update\ndata: {\"project_id\":%q}\n\n", id.String())
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
