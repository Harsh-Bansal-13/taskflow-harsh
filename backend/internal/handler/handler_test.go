package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/broker"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/handler"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/middleware"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/models"
	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/repository"
)

var (
	testPool   *pgxpool.Pool
	testLogger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	jwtSecret  = "test-secret-key"
)

func TestMain(m *testing.M) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		fmt.Println("Skipping integration tests: TEST_DATABASE_URL not set")
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	testPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	defer testPool.Close()

	os.Exit(m.Run())
}

func setupRouter() (*chi.Mux, *repository.UserRepository, *repository.ProjectRepository, *repository.TaskRepository) {
	userRepo := repository.NewUserRepository(testPool)
	projectRepo := repository.NewProjectRepository(testPool)
	taskRepo := repository.NewTaskRepository(testPool)

	authHandler := handler.NewAuthHandler(userRepo, jwtSecret, testLogger)
	evtBroker := broker.New()
	projectHandler := handler.NewProjectHandler(projectRepo, taskRepo, evtBroker, testLogger)
	taskHandler := handler.NewTaskHandler(taskRepo, projectRepo, evtBroker, testLogger)

	r := chi.NewRouter()

	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtSecret))
		r.Get("/projects", projectHandler.List)
		r.Post("/projects", projectHandler.Create)
		r.Get("/projects/{id}", projectHandler.Get)
		r.Patch("/projects/{id}", projectHandler.Update)
		r.Delete("/projects/{id}", projectHandler.Delete)
		r.Get("/projects/{id}/stats", projectHandler.Stats)
		r.Get("/projects/{id}/tasks", taskHandler.List)
		r.Post("/projects/{id}/tasks", taskHandler.Create)
		r.Patch("/tasks/{id}", taskHandler.Update)
		r.Delete("/tasks/{id}", taskHandler.Delete)
	})

	// SSE — auth via ?token= query param
	r.Group(func(r chi.Router) {
		r.Use(middleware.QueryTokenAuth(jwtSecret))
		r.Get("/projects/{id}/events", projectHandler.Events)
	})

	return r, userRepo, projectRepo, taskRepo
}

func cleanupDB() {
	testPool.Exec(context.Background(), "DELETE FROM tasks")
	testPool.Exec(context.Background(), "DELETE FROM projects")
	testPool.Exec(context.Background(), "DELETE FROM users")
}

func TestRegister(t *testing.T) {
	router, _, _, _ := setupRouter()
	defer cleanupDB()

	body := `{"name":"Test User","email":"register@test.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp models.AuthResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.User.Email != "register@test.com" {
		t.Errorf("expected email register@test.com, got %s", resp.User.Email)
	}
}

func TestRegisterValidation(t *testing.T) {
	router, _, _, _ := setupRouter()
	defer cleanupDB()

	body := `{"name":"","email":"","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var resp models.ErrorResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Fields["name"] == "" || resp.Fields["email"] == "" || resp.Fields["password"] == "" {
		t.Error("expected validation errors for name, email, and password")
	}
}

func TestLoginSuccess(t *testing.T) {
	router, _, _, _ := setupRouter()
	defer cleanupDB()

	// Register first
	regBody := `{"name":"Login User","email":"login@test.com","password":"password123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	router.ServeHTTP(regRec, regReq)

	if regRec.Code != http.StatusCreated {
		t.Fatalf("register failed: %d %s", regRec.Code, regRec.Body.String())
	}

	// Login
	loginBody := `{"email":"login@test.com","password":"password123"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", loginRec.Code, loginRec.Body.String())
		return
	}

	var resp models.AuthResponse
	json.NewDecoder(loginRec.Body).Decode(&resp)

	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	router, _, _, _ := setupRouter()
	defer cleanupDB()

	// Register
	regBody := `{"name":"Bad Pass","email":"badpass@test.com","password":"password123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	router.ServeHTTP(regRec, regReq)

	// Login with wrong password
	loginBody := `{"email":"badpass@test.com","password":"wrongpassword"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", loginRec.Code)
	}
}

func TestProtectedEndpointWithoutToken(t *testing.T) {
	router, _, _, _ := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestCreateAndListProjects(t *testing.T) {
	router, _, _, _ := setupRouter()
	defer cleanupDB()

	// Register and get token
	regBody := `{"name":"Project User","email":"project@test.com","password":"password123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	router.ServeHTTP(regRec, regReq)

	var auth models.AuthResponse
	json.NewDecoder(regRec.Body).Decode(&auth)

	// Create project
	createBody := `{"name":"Test Project","description":"A test project"}`
	createReq := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+auth.Token)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", createRec.Code, createRec.Body.String())
		return
	}

	// List projects
	listReq := httptest.NewRequest(http.MethodGet, "/projects", nil)
	listReq.Header.Set("Authorization", "Bearer "+auth.Token)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", listRec.Code, listRec.Body.String())
	}
}

// helper: register a user and return token + user ID
func registerUser(t *testing.T, router *chi.Mux, name, email, password string) models.AuthResponse {
	t.Helper()
	body := fmt.Sprintf(`{"name":%q,"email":%q,"password":%q}`, name, email, password)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("registerUser: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp models.AuthResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	return resp
}

// helper: create a project and return its ID
func createProject(t *testing.T, router *chi.Mux, token, name string) models.Project {
	t.Helper()
	body := fmt.Sprintf(`{"name":%q}`, name)
	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("createProject: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var p models.Project
	json.NewDecoder(rec.Body).Decode(&p)
	return p
}

func TestTaskCRUD(t *testing.T) {
	router, _, _, _ := setupRouter()
	defer cleanupDB()

	auth := registerUser(t, router, "Task User", "taskuser@test.com", "password123")
	project := createProject(t, router, auth.Token, "Task Project")
	projectURL := fmt.Sprintf("/projects/%s/tasks", project.ID)

	// Create task
	createBody := `{"title":"My Task","priority":"high"}`
	createReq := httptest.NewRequest(http.MethodPost, projectURL, bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+auth.Token)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create task: expected 201, got %d: %s", createRec.Code, createRec.Body.String())
	}

	var task models.Task
	json.NewDecoder(createRec.Body).Decode(&task)
	if task.Title != "My Task" {
		t.Errorf("expected title 'My Task', got %q", task.Title)
	}
	if task.Priority != "high" {
		t.Errorf("expected priority 'high', got %q", task.Priority)
	}

	// List tasks
	listReq := httptest.NewRequest(http.MethodGet, projectURL, nil)
	listReq.Header.Set("Authorization", "Bearer "+auth.Token)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Errorf("list tasks: expected 200, got %d", listRec.Code)
	}

	// Update task status
	updateBody := `{"status":"in_progress"}`
	updateReq := httptest.NewRequest(http.MethodPatch, "/tasks/"+task.ID.String(), bytes.NewBufferString(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+auth.Token)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Errorf("update task: expected 200, got %d: %s", updateRec.Code, updateRec.Body.String())
	}

	var updated models.Task
	json.NewDecoder(updateRec.Body).Decode(&updated)
	if updated.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got %q", updated.Status)
	}

	// Delete task
	deleteReq := httptest.NewRequest(http.MethodDelete, "/tasks/"+task.ID.String(), nil)
	deleteReq.Header.Set("Authorization", "Bearer "+auth.Token)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusNoContent {
		t.Errorf("delete task: expected 204, got %d", deleteRec.Code)
	}
}

func TestProjectAuthorizationEdgeCases(t *testing.T) {
	router, _, _, _ := setupRouter()
	defer cleanupDB()

	// Two users
	owner := registerUser(t, router, "Owner", "owner@test.com", "password123")
	stranger := registerUser(t, router, "Stranger", "stranger@test.com", "password123")

	project := createProject(t, router, owner.Token, "Private Project")
	projectURL := "/projects/" + project.ID.String()

	// Stranger cannot GET the project (not a member)
	getReq := httptest.NewRequest(http.MethodGet, projectURL, nil)
	getReq.Header.Set("Authorization", "Bearer "+stranger.Token)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusForbidden {
		t.Errorf("non-member GET project: expected 403, got %d", getRec.Code)
	}

	// Stranger cannot delete the owner's project
	deleteReq := httptest.NewRequest(http.MethodDelete, projectURL, nil)
	deleteReq.Header.Set("Authorization", "Bearer "+stranger.Token)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusForbidden {
		t.Errorf("non-owner DELETE project: expected 403, got %d", deleteRec.Code)
	}

	// Owner can still GET their own project
	ownerGetReq := httptest.NewRequest(http.MethodGet, projectURL, nil)
	ownerGetReq.Header.Set("Authorization", "Bearer "+owner.Token)
	ownerGetRec := httptest.NewRecorder()
	router.ServeHTTP(ownerGetRec, ownerGetReq)
	if ownerGetRec.Code != http.StatusOK {
		t.Errorf("owner GET project: expected 200, got %d", ownerGetRec.Code)
	}
}

func TestTaskValidation(t *testing.T) {
	router, _, _, _ := setupRouter()
	defer cleanupDB()

	auth := registerUser(t, router, "Val User", "val@test.com", "password123")
	project := createProject(t, router, auth.Token, "Val Project")
	projectURL := fmt.Sprintf("/projects/%s/tasks", project.ID)

	// Empty title should fail
	createBody := `{"title":"","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, projectURL, bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+auth.Token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("empty title: expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTaskListFilterByStatus(t *testing.T) {
	router, _, _, _ := setupRouter()
	defer cleanupDB()

	auth := registerUser(t, router, "Filter User", "filter@test.com", "password123")
	project := createProject(t, router, auth.Token, "Filter Project")
	projectURL := fmt.Sprintf("/projects/%s/tasks", project.ID)

	// Create two tasks with different priorities / statuses
	for _, body := range []string{
		`{"title":"Todo Task","priority":"low"}`,
		`{"title":"High Task","priority":"high"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, projectURL, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+auth.Token)
		router.ServeHTTP(httptest.NewRecorder(), req)
	}

	// Filter by status=todo — should return both (default status is todo)
	listReq := httptest.NewRequest(http.MethodGet, projectURL+"?status=todo", nil)
	listReq.Header.Set("Authorization", "Bearer "+auth.Token)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("filter status=todo: expected 200, got %d", listRec.Code)
	}

	var resp models.PaginatedResponse
	if err := json.NewDecoder(listRec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.TotalCount != 2 {
		t.Errorf("filter status=todo: expected 2 tasks, got %d", resp.TotalCount)
	}

	// Invalid status should return 400
	badReq := httptest.NewRequest(http.MethodGet, projectURL+"?status=invalid", nil)
	badReq.Header.Set("Authorization", "Bearer "+auth.Token)
	badRec := httptest.NewRecorder()
	router.ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusBadRequest {
		t.Errorf("invalid status: expected 400, got %d", badRec.Code)
	}
}

func TestProjectListPagination(t *testing.T) {
	router, _, _, _ := setupRouter()
	defer cleanupDB()

	auth := registerUser(t, router, "Paginate User", "paginate@test.com", "password123")

	// Create 3 projects
	for i := 1; i <= 3; i++ {
		createProject(t, router, auth.Token, fmt.Sprintf("Project %d", i))
	}

	// page=1 limit=2 — should return 2 with total_count=3
	listReq := httptest.NewRequest(http.MethodGet, "/projects?page=1&limit=2", nil)
	listReq.Header.Set("Authorization", "Bearer "+auth.Token)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("pagination: expected 200, got %d", listRec.Code)
	}

	var resp models.PaginatedResponse
	if err := json.NewDecoder(listRec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.TotalCount != 3 {
		t.Errorf("pagination: expected total_count=3, got %d", resp.TotalCount)
	}
	if resp.TotalPages != 2 {
		t.Errorf("pagination: expected total_pages=2, got %d", resp.TotalPages)
	}
	if resp.Page != 1 {
		t.Errorf("pagination: expected page=1, got %d", resp.Page)
	}
}

func TestSSEConnectionAndAuth(t *testing.T) {
	router, _, _, _ := setupRouter()
	defer cleanupDB()

	auth := registerUser(t, router, "SSE User", "sse@test.com", "password123")
	project := createProject(t, router, auth.Token, "SSE Project")
	eventsURL := fmt.Sprintf("/projects/%s/events", project.ID)

	// No token — should be 401
	unauthReq := httptest.NewRequest(http.MethodGet, eventsURL, nil)
	unauthRec := httptest.NewRecorder()
	router.ServeHTTP(unauthRec, unauthReq)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Errorf("SSE no token: expected 401, got %d", unauthRec.Code)
	}

	// Wrong token — should be 401
	wrongReq := httptest.NewRequest(http.MethodGet, eventsURL+"?token=bad.token.here", nil)
	wrongRec := httptest.NewRecorder()
	router.ServeHTTP(wrongRec, wrongReq)
	if wrongRec.Code != http.StatusUnauthorized {
		t.Errorf("SSE wrong token: expected 401, got %d", wrongRec.Code)
	}

	// Valid token — should open SSE stream (200 text/event-stream)
	validReq := httptest.NewRequest(http.MethodGet, eventsURL+"?token="+auth.Token, nil)
	validReq.Header.Set("Accept", "text/event-stream")
	validRec := httptest.NewRecorder()
	// Use a cancelled context so the handler returns promptly
	ctx, cancel := context.WithCancel(validReq.Context())
	cancel()
	router.ServeHTTP(validRec, validReq.WithContext(ctx))
	if validRec.Code != http.StatusOK {
		t.Errorf("SSE valid token: expected 200, got %d: %s", validRec.Code, validRec.Body.String())
	}
	if ct := validRec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("SSE content-type: expected text/event-stream, got %q", ct)
	}
}
