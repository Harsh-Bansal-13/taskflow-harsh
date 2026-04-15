import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '@/context/AuthContext';
import { useProjectDetail } from '@/hooks/useProjectDetail';
import { Button } from '@/components/ui/Button';
import { Card, CardContent } from '@/components/ui/Card';
import { Modal } from '@/components/Modal';
import { Spinner } from '@/components/ui/Spinner';
import { TaskCard } from '@/components/TaskCard';
import { TaskForm } from '@/components/TaskForm';
import { KanbanBoard } from '@/components/KanbanBoard';
import { TaskFilterBar } from '@/components/TaskFilterBar';
import { EditProjectModal } from '@/components/EditProjectModal';
import { ConfirmDeleteModal } from '@/components/ConfirmDeleteModal';
import { ArrowLeft, Plus, Pencil, Trash2, AlertCircle, ListTodo } from 'lucide-react';

export default function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const h = useProjectDetail(id, user?.id);

  if (h.loading) return <Spinner className="mt-20" />;

  if (h.error || !h.project) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="bg-destructive/10 text-destructive p-4 rounded-md text-center">
          <AlertCircle className="h-6 w-6 mx-auto mb-2" />
          {h.error || 'Project not found'}
          <div className="mt-4">
            <Button variant="outline" onClick={() => navigate('/')}>
              Back to Projects
            </Button>
          </div>
        </div>
      </div>
    );
  }

  const handleDeleteProject = async () => {
    const navigateAway = await h.handleDeleteProject();
    if (navigateAway) navigate('/');
  };

  return (
    <div className="max-w-5xl mx-auto px-4 py-8">
      {/* Action error banner */}
      {h.actionError && (
        <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-md mb-4 flex items-center justify-between">
          <span>{h.actionError}</span>
          <Button variant="ghost" size="sm" onClick={() => h.setActionError('')}>
            Dismiss
          </Button>
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center gap-4 mb-6">
        <Button variant="ghost" size="sm" onClick={() => navigate('/')}>
          <ArrowLeft className="h-4 w-4 mr-1" /> Back
        </Button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold">{h.project.name}</h1>
          {h.project.description && (
            <p className="text-muted-foreground mt-1">{h.project.description}</p>
          )}
        </div>
        {h.isOwner && (
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={h.openEditProject}>
              <Pencil className="h-4 w-4 mr-1" /> Edit
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => h.setShowDeleteConfirm(true)}
            >
              <Trash2 className="h-4 w-4 mr-1" /> Delete
            </Button>
          </div>
        )}
      </div>

      {/* Filters + Add */}
      <TaskFilterBar
        tasks={h.project.tasks}
        statusFilter={h.statusFilter}
        onStatusFilterChange={h.setStatusFilter}
        assigneeFilter={h.assigneeFilter}
        onAssigneeFilterChange={h.setAssigneeFilter}
        assigneeIds={h.assigneeIds}
        assigneeMap={h.assigneeMap}
        currentUserId={user?.id}
        view={h.view}
        onViewChange={h.setView}
        onAddTask={h.openCreateTask}
      />

      {/* Tasks */}
      {h.view === 'board' ? (
        <KanbanBoard
          tasks={h.filteredTasks}
          currentUserId={user?.id}
          onStatusChange={h.handleStatusChange}
          onEdit={h.openEditTask}
          onDelete={h.confirmDeleteTask}
        />
      ) : h.filteredTasks.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <ListTodo className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground text-lg mb-2">
              {h.statusFilter || h.assigneeFilter ? 'No tasks match this filter' : 'No tasks yet'}
            </p>
            {!h.statusFilter && !h.assigneeFilter && (
              <Button size="sm" onClick={h.openCreateTask} className="mt-2">
                <Plus className="h-4 w-4 mr-1" /> Create first task
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {h.filteredTasks.map((task) => (
            <Card key={task.id} className="hover:border-primary/30 transition-colors">
              <CardContent className="p-4">
                <TaskCard
                  task={task}
                  currentUserId={user?.id}
                  onStatusChange={h.handleStatusChange}
                  onEdit={h.openEditTask}
                  onDelete={h.confirmDeleteTask}
                />
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Task Modal */}
      <Modal
        isOpen={h.showTaskModal}
        onClose={() => h.setShowTaskModal(false)}
        title={h.editingTask ? 'Edit Task' : 'Create Task'}
      >
        <TaskForm
          form={h.taskForm}
          onChange={h.setTaskForm}
          onSubmit={h.handleTaskSubmit}
          onCancel={() => h.setShowTaskModal(false)}
          isEditing={!!h.editingTask}
          saving={h.taskSaving}
          error={h.taskError}
          users={h.users}
        />
      </Modal>

      {/* Edit Project Modal */}
      <EditProjectModal
        isOpen={h.showEditProject}
        onClose={() => h.setShowEditProject(false)}
        name={h.editProjectName}
        onNameChange={h.setEditProjectName}
        description={h.editProjectDesc}
        onDescriptionChange={h.setEditProjectDesc}
        saving={h.editProjectSaving}
        onSubmit={h.handleEditProject}
      />

      {/* Delete Project Confirm */}
      <ConfirmDeleteModal
        isOpen={h.showDeleteConfirm}
        onClose={() => h.setShowDeleteConfirm(false)}
        title="Delete Project"
        message={`This will permanently delete "${h.project.name}" and all its tasks. This action cannot be undone.`}
        onConfirm={handleDeleteProject}
        deleting={h.deleting}
        confirmLabel="Delete Project"
      />

      {/* Delete Task Confirm */}
      <ConfirmDeleteModal
        isOpen={h.showDeleteTaskConfirm}
        onClose={() => { h.setShowDeleteTaskConfirm(false); h.setDeletingTaskId(null); }}
        title="Delete Task"
        message="Are you sure you want to delete this task? This action cannot be undone."
        onConfirm={h.handleDeleteTask}
        deleting={h.deletingTask}
        confirmLabel="Delete Task"
      />
    </div>
  );
}
