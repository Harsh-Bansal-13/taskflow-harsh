import { useState } from 'react';
import { useProjectData } from './useProjectData';
import { useTaskFilters } from './useTaskFilters';
import { useTaskModal } from './useTaskModal';
import { useProjectModal } from './useProjectModal';

export function useProjectDetail(id: string | undefined, currentUserId: string | undefined) {
  const { project, setProject, users, loading, error, fetchProject } = useProjectData(id);
  const filters = useTaskFilters(project?.tasks || [], users);
  const tasks = useTaskModal(id, project, setProject, fetchProject);
  const proj = useProjectModal(id, project, fetchProject);
  const [view, setView] = useState<'list' | 'board'>('list');

  const isOwner = project?.owner_id === currentUserId;

  // Merge action errors from both sub-hooks
  const actionError = tasks.actionError || proj.actionError;
  const setActionError = (msg: string) => {
    tasks.setActionError(msg);
    proj.setActionError(msg);
  };

  return {
    // Data
    project,
    users,
    loading,
    error,
    actionError,
    isOwner,
    assigneeMap: filters.assigneeMap,
    assigneeIds: filters.assigneeIds,
    filteredTasks: filters.filteredTasks,

    // Filters
    statusFilter: filters.statusFilter,
    setStatusFilter: filters.setStatusFilter,
    assigneeFilter: filters.assigneeFilter,
    setAssigneeFilter: filters.setAssigneeFilter,

    // Task modal
    showTaskModal: tasks.showTaskModal,
    setShowTaskModal: tasks.setShowTaskModal,
    editingTask: tasks.editingTask,
    taskForm: tasks.taskForm,
    setTaskForm: tasks.setTaskForm,
    taskSaving: tasks.taskSaving,
    taskError: tasks.taskError,

    // Task actions
    openCreateTask: tasks.openCreateTask,
    openEditTask: tasks.openEditTask,
    handleTaskSubmit: tasks.handleTaskSubmit,
    handleStatusChange: tasks.handleStatusChange,
    confirmDeleteTask: tasks.confirmDeleteTask,
    handleDeleteTask: tasks.handleDeleteTask,
    showDeleteTaskConfirm: tasks.showDeleteTaskConfirm,
    setShowDeleteTaskConfirm: tasks.setShowDeleteTaskConfirm,
    deletingTaskId: tasks.deletingTaskId,
    setDeletingTaskId: tasks.setDeletingTaskId,
    deletingTask: tasks.deletingTask,

    // Project edit
    showEditProject: proj.showEditProject,
    setShowEditProject: proj.setShowEditProject,
    editProjectName: proj.editProjectName,
    setEditProjectName: proj.setEditProjectName,
    editProjectDesc: proj.editProjectDesc,
    setEditProjectDesc: proj.setEditProjectDesc,
    editProjectSaving: proj.editProjectSaving,
    openEditProject: proj.openEditProject,
    handleEditProject: proj.handleEditProject,

    // Project delete
    showDeleteConfirm: proj.showDeleteConfirm,
    setShowDeleteConfirm: proj.setShowDeleteConfirm,
    deleting: proj.deleting,
    handleDeleteProject: proj.handleDeleteProject,

    // View
    view,
    setView,

    // Misc
    setActionError,
  };
}
