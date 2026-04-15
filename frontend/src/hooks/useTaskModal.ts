import { useState, type FormEvent } from 'react';
import { createTask, updateTask, deleteTask } from '@/api/endpoints';
import type { Task, ProjectWithTasks } from '@/types';
import type { TaskFormData } from '@/components/TaskForm';

const EMPTY_FORM: TaskFormData = {
  title: '',
  description: '',
  priority: 'medium',
  assignee_id: '',
  due_date: '',
  status: 'todo',
};

export function useTaskModal(
  projectId: string | undefined,
  project: ProjectWithTasks | null,
  setProject: (p: ProjectWithTasks | null) => void,
  fetchProject: () => Promise<void>,
) {
  const [showTaskModal, setShowTaskModal] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [taskForm, setTaskForm] = useState<TaskFormData>(EMPTY_FORM);
  const [taskSaving, setTaskSaving] = useState(false);
  const [taskError, setTaskError] = useState('');

  // Delete task confirm
  const [showDeleteTaskConfirm, setShowDeleteTaskConfirm] = useState(false);
  const [deletingTaskId, setDeletingTaskId] = useState<string | null>(null);
  const [deletingTask, setDeletingTask] = useState(false);
  const [actionError, setActionError] = useState('');

  // Optimistic status change
  const handleStatusChange = async (task: Task, newStatus: string) => {
    if (!project) return;
    const oldTasks = [...project.tasks];
    setProject({
      ...project,
      tasks: project.tasks.map((t) =>
        t.id === task.id ? { ...t, status: newStatus as Task['status'] } : t
      ),
    });
    try {
      await updateTask(task.id, { status: newStatus });
    } catch {
      setProject({ ...project, tasks: oldTasks });
    }
  };

  const openCreateTask = () => {
    setEditingTask(null);
    setTaskForm(EMPTY_FORM);
    setTaskError('');
    setShowTaskModal(true);
  };

  const openEditTask = (task: Task) => {
    setEditingTask(task);
    setTaskForm({
      title: task.title,
      description: task.description || '',
      priority: task.priority,
      assignee_id: task.assignee_id || '',
      due_date: task.due_date || '',
      status: task.status,
    });
    setTaskError('');
    setShowTaskModal(true);
  };

  const handleTaskSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!taskForm.title.trim()) {
      setTaskError('Title is required');
      return;
    }
    setTaskSaving(true);
    setTaskError('');
    try {
      if (editingTask) {
        await updateTask(editingTask.id, {
          title: taskForm.title,
          description: taskForm.description || undefined,
          status: taskForm.status,
          priority: taskForm.priority,
          assignee_id: taskForm.assignee_id || undefined,
          due_date: taskForm.due_date || undefined,
        });
      } else {
        await createTask(projectId!, {
          title: taskForm.title,
          description: taskForm.description || undefined,
          priority: taskForm.priority,
          assignee_id: taskForm.assignee_id || undefined,
          due_date: taskForm.due_date || undefined,
        });
      }
      setShowTaskModal(false);
      fetchProject();
    } catch (err: any) {
      setTaskError(err.response?.data?.error || 'Failed to save task');
    } finally {
      setTaskSaving(false);
    }
  };

  const confirmDeleteTask = (taskId: string) => {
    setDeletingTaskId(taskId);
    setShowDeleteTaskConfirm(true);
  };

  const handleDeleteTask = async () => {
    if (!deletingTaskId) return;
    setDeletingTask(true);
    try {
      await deleteTask(deletingTaskId);
      setShowDeleteTaskConfirm(false);
      setDeletingTaskId(null);
      fetchProject();
    } catch {
      setActionError('Failed to delete task');
      setShowDeleteTaskConfirm(false);
    } finally {
      setDeletingTask(false);
    }
  };

  return {
    showTaskModal,
    setShowTaskModal,
    editingTask,
    taskForm,
    setTaskForm,
    taskSaving,
    taskError,
    openCreateTask,
    openEditTask,
    handleTaskSubmit,
    handleStatusChange,
    confirmDeleteTask,
    handleDeleteTask,
    showDeleteTaskConfirm,
    setShowDeleteTaskConfirm,
    deletingTaskId,
    setDeletingTaskId,
    deletingTask,
    actionError,
    setActionError,
  };
}
