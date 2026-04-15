package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/models"
)

type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, t *models.Task) error {
	query := `
		INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, created_by, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`
	return r.db.QueryRow(ctx, query,
		t.ID, t.Title, t.Description, t.Status, t.Priority,
		t.ProjectID, t.AssigneeID, t.CreatedBy, t.DueDate,
	).Scan(&t.CreatedAt, &t.UpdatedAt)
}

func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	t := &models.Task{}
	query := `
		SELECT t.id, t.title, t.description, t.status, t.priority, t.project_id, t.assignee_id, u.name, t.created_by, t.due_date::text, t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN users u ON u.id = t.assignee_id
		WHERE t.id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.ProjectID, &t.AssigneeID, &t.AssigneeName, &t.CreatedBy, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}

func (r *TaskRepository) ListByProject(ctx context.Context, projectID uuid.UUID, status, assignee string, page, limit int) ([]models.Task, int, error) {
	conditions := []string{"t.project_id = $1"}
	args := []interface{}{projectID}
	argIdx := 2

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if assignee != "" {
		conditions = append(conditions, fmt.Sprintf("t.assignee_id = $%d", argIdx))
		args = append(args, assignee)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks t WHERE %s", where)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	// Data
	offset := (page - 1) * limit
	dataQuery := fmt.Sprintf(`
		SELECT t.id, t.title, t.description, t.status, t.priority, t.project_id, t.assignee_id, u.name, t.created_by, t.due_date::text, t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN users u ON u.id = t.assignee_id
		WHERE %s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.AssigneeName, &t.CreatedBy, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, total, nil
}

// GetAllByProject returns every task for a project without a row cap.
// Used by the project detail endpoint so projects with >1000 tasks are not silently truncated.
func (r *TaskRepository) GetAllByProject(ctx context.Context, projectID uuid.UUID) ([]models.Task, error) {
	query := `
		SELECT t.id, t.title, t.description, t.status, t.priority, t.project_id, t.assignee_id, u.name, t.created_by, t.due_date::text, t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN users u ON u.id = t.assignee_id
		WHERE t.project_id = $1
		ORDER BY t.created_at DESC`
	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("get all tasks: %w", err)
	}
	defer rows.Close()
	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.AssigneeName, &t.CreatedBy, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, id uuid.UUID, req models.UpdateTaskRequest) (*models.Task, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *req.Title)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *req.Status)
		argIdx++
	}
	if req.Priority != nil {
		setClauses = append(setClauses, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, *req.Priority)
		argIdx++
	}
	if req.AssigneeID != nil {
		if *req.AssigneeID == "" {
			setClauses = append(setClauses, fmt.Sprintf("assignee_id = $%d", argIdx))
			args = append(args, nil)
		} else {
			setClauses = append(setClauses, fmt.Sprintf("assignee_id = $%d", argIdx))
			parsed, err := uuid.Parse(*req.AssigneeID)
			if err != nil {
				return nil, fmt.Errorf("invalid assignee_id: %w", err)
			}
			args = append(args, parsed)
		}
		argIdx++
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			setClauses = append(setClauses, fmt.Sprintf("due_date = $%d", argIdx))
			args = append(args, nil)
		} else {
			setClauses = append(setClauses, fmt.Sprintf("due_date = $%d", argIdx))
			args = append(args, *req.DueDate)
		}
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	setStr := strings.Join(setClauses, ", ")
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE tasks SET %s WHERE id = $%d
		RETURNING id, title, description, status, priority, project_id, assignee_id,
		  (SELECT name FROM users WHERE users.id = tasks.assignee_id),
		  created_by, due_date::text, created_at, updated_at`,
		setStr, argIdx)

	t := &models.Task{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.ProjectID, &t.AssigneeID, &t.AssigneeName, &t.CreatedBy, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("update task: %w", err)
	}
	return t, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	return err
}

func (r *TaskRepository) GetProjectStats(ctx context.Context, projectID uuid.UUID) (*models.ProjectStats, error) {
	stats := &models.ProjectStats{
		StatusCounts:   make(map[string]int),
		AssigneeCounts: make(map[string]int),
	}

	// Status counts
	rows, err := r.db.Query(ctx, `SELECT status, COUNT(*) FROM tasks WHERE project_id = $1 GROUP BY status`, projectID)
	if err != nil {
		return nil, fmt.Errorf("status counts: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats.StatusCounts[status] = count
	}

	// Assignee counts
	rows2, err := r.db.Query(ctx, `
		SELECT COALESCE(u.name, 'Unassigned'), COUNT(*)
		FROM tasks t
		LEFT JOIN users u ON t.assignee_id = u.id
		WHERE t.project_id = $1
		GROUP BY u.id, u.name`, projectID)
	if err != nil {
		return nil, fmt.Errorf("assignee counts: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var name string
		var count int
		if err := rows2.Scan(&name, &count); err != nil {
			return nil, err
		}
		stats.AssigneeCounts[name] = count
	}

	return stats, nil
}
