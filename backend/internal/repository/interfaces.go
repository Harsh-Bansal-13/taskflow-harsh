package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/models"
)

// UserRepo defines the interface for user data access.
type UserRepo interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	ListAll(ctx context.Context) ([]models.User, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.User, error)
}

// ProjectRepo defines the interface for project data access.
type ProjectRepo interface {
	Create(ctx context.Context, p *models.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error)
	IsMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.Project, int, error)
	Update(ctx context.Context, id uuid.UUID, name *string, description *string) (*models.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// TaskRepo defines the interface for task data access.
type TaskRepo interface {
	Create(ctx context.Context, t *models.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, status, assignee string, page, limit int) ([]models.Task, int, error)
	GetAllByProject(ctx context.Context, projectID uuid.UUID) ([]models.Task, error)
	Update(ctx context.Context, id uuid.UUID, req models.UpdateTaskRequest) (*models.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetProjectStats(ctx context.Context, projectID uuid.UUID) (*models.ProjectStats, error)
}
