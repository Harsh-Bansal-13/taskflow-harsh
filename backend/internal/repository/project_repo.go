package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/models"
)

type ProjectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, p *models.Project) error {
	query := `INSERT INTO projects (id, name, description, owner_id) VALUES ($1, $2, $3, $4) RETURNING created_at`
	return r.db.QueryRow(ctx, query, p.ID, p.Name, p.Description, p.OwnerID).Scan(&p.CreatedAt)
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	p := &models.Project{}
	query := `SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	return p, nil
}

// IsMember checks whether the user owns the project or has tasks (assigned or created) in it.
func (r *ProjectRepository) IsMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM projects WHERE id = $1 AND owner_id = $2
			UNION ALL
			SELECT 1 FROM tasks WHERE project_id = $1 AND (assignee_id = $2 OR created_by = $2)
		)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, projectID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check membership: %w", err)
	}
	return exists, nil
}

func (r *ProjectRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.Project, int, error) {
	countQuery := `
		SELECT COUNT(DISTINCT p.id)
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1 OR t.created_by = $1`

	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}

	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1 OR t.created_by = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`

	offset := (page - 1) * limit
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, total, nil
}

func (r *ProjectRepository) Update(ctx context.Context, id uuid.UUID, name *string, description *string) (*models.Project, error) {
	p := &models.Project{}
	query := `
		UPDATE projects 
		SET name = COALESCE($2, name), description = COALESCE($3, description)
		WHERE id = $1
		RETURNING id, name, description, owner_id, created_at`
	err := r.db.QueryRow(ctx, query, id, name, description).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("update project: %w", err)
	}
	return p, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	return err
}
