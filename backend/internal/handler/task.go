package handler

import (
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

type TaskHandler struct {
	taskRepo    repository.TaskRepo
	projectRepo repository.ProjectRepo
	broker      *broker.EventBroker
	logger      *slog.Logger
}

func NewTaskHandler(taskRepo repository.TaskRepo, projectRepo repository.ProjectRepo, b *broker.EventBroker, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{taskRepo: taskRepo, projectRepo: projectRepo, broker: b, logger: logger}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id", nil)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil {
		h.logger.Error("list tasks: get project", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if project == nil {
		writeError(w, http.StatusNotFound, "not found", nil)
		return
	}

	status := r.URL.Query().Get("status")
	assignee := r.URL.Query().Get("assignee")
	page, limit := parsePagination(r)

	// Validate status filter
	if status != "" {
		validStatuses := map[string]bool{"todo": true, "in_progress": true, "done": true}
		if !validStatuses[status] {
			writeError(w, http.StatusBadRequest, "validation failed", map[string]string{"status": "must be todo, in_progress, or done"})
			return
		}
	}

	// Validate assignee filter
	if assignee != "" {
		if _, err := uuid.Parse(assignee); err != nil {
			writeError(w, http.StatusBadRequest, "validation failed", map[string]string{"assignee": "must be a valid UUID"})
			return
		}
	}

	tasks, total, err := h.taskRepo.ListByProject(r.Context(), projectID, status, assignee, page, limit)
	if err != nil {
		h.logger.Error("list tasks", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if tasks == nil {
		tasks = []models.Task{}
	}

	writeJSON(w, http.StatusOK, paginatedResponse(tasks, page, limit, total))
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id", nil)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil {
		h.logger.Error("create task: get project", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if project == nil {
		writeError(w, http.StatusNotFound, "not found", nil)
		return
	}

	userID := middleware.GetUserID(r.Context())

	var req models.CreateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "validation failed", map[string]string{"title": "is required"})
		return
	}

	priority := "medium"
	if req.Priority != nil {
		validPriorities := map[string]bool{"low": true, "medium": true, "high": true}
		if !validPriorities[*req.Priority] {
			writeError(w, http.StatusBadRequest, "validation failed", map[string]string{"priority": "must be low, medium, or high"})
			return
		}
		priority = *req.Priority
	}

	var assigneeID *uuid.UUID
	if req.AssigneeID != nil && *req.AssigneeID != "" {
		parsed, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "validation failed", map[string]string{"assignee_id": "invalid uuid"})
			return
		}
		assigneeID = &parsed
	}

	task := &models.Task{
		ID:          uuid.New(),
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		Status:      "todo",
		Priority:    priority,
		ProjectID:   projectID,
		AssigneeID:  assigneeID,
		CreatedBy:   userID,
		DueDate:     req.DueDate,
	}

	if err := h.taskRepo.Create(r.Context(), task); err != nil {
		h.logger.Error("create task", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	writeJSON(w, http.StatusCreated, task)
	h.broker.Publish(projectID)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id", nil)
		return
	}

	task, err := h.taskRepo.GetByID(r.Context(), taskID)
	if err != nil {
		h.logger.Error("update task: get", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "not found", nil)
		return
	}

	// Authorization: project owner, task creator, or task assignee can update
	userID := middleware.GetUserID(r.Context())
	project, err := h.projectRepo.GetByID(r.Context(), task.ProjectID)
	if err != nil {
		h.logger.Error("update task: get project", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID
	if project.OwnerID != userID && task.CreatedBy != userID && !isAssignee {
		writeError(w, http.StatusForbidden, "forbidden", nil)
		return
	}

	var req models.UpdateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := map[string]bool{"todo": true, "in_progress": true, "done": true}
		if !validStatuses[*req.Status] {
			writeError(w, http.StatusBadRequest, "validation failed", map[string]string{"status": "must be todo, in_progress, or done"})
			return
		}
	}

	// Validate priority if provided
	if req.Priority != nil {
		validPriorities := map[string]bool{"low": true, "medium": true, "high": true}
		if !validPriorities[*req.Priority] {
			writeError(w, http.StatusBadRequest, "validation failed", map[string]string{"priority": "must be low, medium, or high"})
			return
		}
	}

	updated, err := h.taskRepo.Update(r.Context(), taskID, req)
	if err != nil {
		h.logger.Error("update task", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	writeJSON(w, http.StatusOK, updated)
	h.broker.Publish(task.ProjectID)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id", nil)
		return
	}

	task, err := h.taskRepo.GetByID(r.Context(), taskID)
	if err != nil {
		h.logger.Error("delete task: get", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "not found", nil)
		return
	}

	userID := middleware.GetUserID(r.Context())

	// Check: project owner or task creator
	project, err := h.projectRepo.GetByID(r.Context(), task.ProjectID)
	if err != nil {
		h.logger.Error("delete task: get project", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	if project.OwnerID != userID && task.CreatedBy != userID {
		writeError(w, http.StatusForbidden, "forbidden", nil)
		return
	}

	if err := h.taskRepo.Delete(r.Context(), taskID); err != nil {
		h.logger.Error("delete task", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	h.broker.Publish(task.ProjectID)
}
