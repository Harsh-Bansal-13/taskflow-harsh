import { useState, type FormEvent } from 'react';
import { updateProject, deleteProject } from '@/api/endpoints';
import type { ProjectWithTasks } from '@/types';

export function useProjectModal(
  projectId: string | undefined,
  project: ProjectWithTasks | null,
  fetchProject: () => Promise<void>,
) {
  const [showEditProject, setShowEditProject] = useState(false);
  const [editProjectName, setEditProjectName] = useState('');
  const [editProjectDesc, setEditProjectDesc] = useState('');
  const [editProjectSaving, setEditProjectSaving] = useState(false);
  const [actionError, setActionError] = useState('');

  // Delete project confirm
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const openEditProject = () => {
    if (!project) return;
    setEditProjectName(project.name);
    setEditProjectDesc(project.description || '');
    setShowEditProject(true);
  };

  const handleEditProject = async (e: FormEvent) => {
    e.preventDefault();
    setEditProjectSaving(true);
    try {
      await updateProject(projectId!, { name: editProjectName, description: editProjectDesc });
      setShowEditProject(false);
      fetchProject();
    } catch {
      setActionError('Failed to update project');
    } finally {
      setEditProjectSaving(false);
    }
  };

  const handleDeleteProject = async () => {
    setDeleting(true);
    try {
      await deleteProject(projectId!);
      return true; // Signal caller to navigate
    } catch {
      setActionError('Failed to delete project');
      setDeleting(false);
      return false;
    }
  };

  return {
    showEditProject,
    setShowEditProject,
    editProjectName,
    setEditProjectName,
    editProjectDesc,
    setEditProjectDesc,
    editProjectSaving,
    openEditProject,
    handleEditProject,
    showDeleteConfirm,
    setShowDeleteConfirm,
    deleting,
    handleDeleteProject,
    actionError,
    setActionError,
  };
}
