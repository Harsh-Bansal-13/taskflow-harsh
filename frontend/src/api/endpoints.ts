import api from './client';
import type {
  AuthResponse,
  Project,
  ProjectWithTasks,
  Task,
  User,
  PaginatedResponse,
  ProjectStats,
} from '../types';

// Auth
export const register = (name: string, email: string, password: string) =>
  api.post<AuthResponse>('/auth/register', { name, email, password });

export const login = (email: string, password: string) =>
  api.post<AuthResponse>('/auth/login', { email, password });

// Projects
export const getProjects = (page = 1, limit = 20) =>
  api.get<PaginatedResponse<Project>>('/projects', { params: { page, limit } });

export const getProject = (id: string) =>
  api.get<ProjectWithTasks>(`/projects/${id}`);

export const createProject = (name: string, description?: string) =>
  api.post<Project>('/projects', { name, description });

export const updateProject = (id: string, data: { name?: string; description?: string }) =>
  api.patch<Project>(`/projects/${id}`, data);

export const deleteProject = (id: string) =>
  api.delete(`/projects/${id}`);

export const getProjectStats = (id: string) =>
  api.get<ProjectStats>(`/projects/${id}/stats`);

// Users
export const getUsers = (projectId?: string) =>
  api.get<User[]>('/users', { params: projectId ? { project_id: projectId } : undefined });

// Tasks
export const getProjectTasks = (
  projectId: string,
  params?: { status?: string; assignee?: string; page?: number; limit?: number }
) =>
  api.get<PaginatedResponse<Task>>(`/projects/${projectId}/tasks`, { params });

export const createTask = (
  projectId: string,
  data: {
    title: string;
    description?: string;
    priority?: string;
    assignee_id?: string;
    due_date?: string;
  }
) => api.post<Task>(`/projects/${projectId}/tasks`, data);

export const updateTask = (
  id: string,
  data: {
    title?: string;
    description?: string;
    status?: string;
    priority?: string;
    assignee_id?: string;
    due_date?: string;
  }
) => api.patch<Task>(`/tasks/${id}`, data);

export const deleteTask = (id: string) =>
  api.delete(`/tasks/${id}`);
