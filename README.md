# 🚀 TaskFlow — Scalable Real-Time Task Management System

TaskFlow is a full-stack project management platform designed to simulate how real-world systems are built. It supports project organization, task tracking, and real-time collaboration with a strong focus on backend design, performance, and clean architecture.

This project is not just a CRUD app — it demonstrates **system design thinking, real-time data flow, and production-oriented engineering decisions**.

---

## 🧠 Problem Statement

Most simple task managers fail in one of two ways:
- Either they lack real-time synchronization
- Or they are over-engineered and hard to maintain

TaskFlow is built to strike the balance:
- Simple enough to reason about
- Powerful enough to reflect real-world workflows

---

## ⚙️ Core Capabilities

- Multi-user authentication system
- Project-based task organization
- Task lifecycle management (todo → in_progress → done)
- Real-time updates across sessions
- Kanban-style drag-and-drop interface
- Pagination and filtering for scalability
- Project-level analytics

---

## 🏗️ System Architecture

### High-Level Flow

```
Client (React)
     ↓
API Layer (Go - Chi Router)
     ↓
Service Layer (Business Logic)
     ↓
Repository Layer (SQL Queries)
     ↓
PostgreSQL Database
```

---

### Backend Architecture

The backend follows a layered architecture:

#### 1. Handler Layer
- Responsible for HTTP request/response
- Validates input
- Extracts auth context

#### 2. Service Layer
- Contains business logic
- Enforces rules (ownership, permissions)

#### 3. Repository Layer
- Handles database queries
- Uses raw SQL for performance

---

### Why This Architecture?

- Separation of concerns → easier testing & scaling  
- Replaceable components → flexible design  
- Clear debugging boundaries  

---

## 🧩 Database Schema (Conceptual)

### Users
```sql
id (uuid)
name
email (unique)
password_hash
created_at
```

### Projects
```sql
id (uuid)
name
description
owner_id (fk → users.id)
created_at
```

### Tasks
```sql
id (uuid)
title
description
status (todo / in_progress / done)
priority (low / medium / high)
assignee_id (nullable)
project_id (fk)
due_date
created_at
```

---

## 🔐 Authentication & Security

- JWT-based authentication
- Stateless backend (no session storage)
- Password hashing (bcrypt)
- Protected routes via middleware

---

## 🔄 Real-Time System (SSE)

Instead of WebSockets, TaskFlow uses **Server-Sent Events (SSE)**.

### Flow:
1. Client subscribes to `/projects/:id/events`
2. Backend maintains connection
3. On task update → event is published
4. All connected clients receive update instantly

### Why SSE?
- Simpler than WebSockets
- Works over HTTP
- Ideal for one-way updates

---

## 📡 API Design

### Authentication

#### POST `/auth/register`
Creates a new user.

#### POST `/auth/login`
Returns JWT token.

---

### Projects

#### GET `/projects`
- Paginated project list
- Only user-owned projects

#### POST `/projects`
- Creates a project

#### GET `/projects/:id`
- Returns project + tasks

#### PATCH `/projects/:id`
- Owner-only update

#### DELETE `/projects/:id`
- Owner-only delete

#### GET `/projects/:id/stats`
Returns:
- Task status distribution
- Task ownership breakdown

---

### Tasks

#### GET `/projects/:id/tasks`
Supports:
- Pagination
- Status filter
- Assignee filter

#### POST `/projects/:id/tasks`
Creates a task.

#### PATCH `/tasks/:id`
Updates:
- status
- priority
- title
- etc.

#### DELETE `/tasks/:id`
Deletes task.

---

### Users

#### GET `/users`
Returns list of users for assignment.

---

## 📦 Example API Response

```json
{
  "data": [
    {
      "id": "task_id",
      "title": "Implement API",
      "status": "in_progress",
      "priority": "high"
    }
  ],
  "page": 1,
  "total_pages": 2
}
```

---

## ⚠️ Error Handling Strategy

All errors follow a consistent structure:

```json
{
  "error": "validation failed",
  "fields": {
    "email": "is required"
  }
}
```

---

## 🎨 Frontend Design

- React + TypeScript
- Component-driven structure
- Tailwind CSS for styling
- Axios with interceptors

### Key Patterns

- Optimistic UI updates  
- Global auth state via Context  
- Reusable UI components  

---

## ⚖️ Engineering Tradeoffs

### 1. No ORM
- ✅ Better performance
- ❌ More manual queries

### 2. No Refresh Tokens
- ✅ Simpler implementation
- ❌ Requires re-login after expiry

### 3. In-Memory SSE Broker
- ✅ Fast and simple
- ❌ Not horizontally scalable

---

## 🚀 Running Locally

```bash
git clone https://github.com/Harsh-Bansal-13/taskflow-harsh
cd taskflow-harsh
cp .env.example .env
docker compose up --build
```

---

## 🔐 Test Credentials

```
Email: test@example.com
Password: password123
```

---

## 🔮 Future Enhancements

- Redis-based pub/sub for scalability
- Refresh token rotation
- Role-based access control (RBAC)
- Distributed rate limiting
- Full test suite (unit + integration + E2E)
- CI/CD pipeline

---

## 💡 Key Learnings

- Designing clean backend architecture in Go  
- Implementing real-time systems without WebSockets  
- Managing state efficiently in React  
- Handling tradeoffs in system design  

---

## 👨‍💻 Author

Harsh Bansal  
Software Engineer | Full Stack Developer  

---

## ⭐ Final Thought

TaskFlow is built with a focus on **how real systems are designed**, not just how features are implemented.
