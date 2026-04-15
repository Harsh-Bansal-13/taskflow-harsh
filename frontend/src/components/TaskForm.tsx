import { type FormEvent } from 'react';
import type { User } from '@/types';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Textarea } from '@/components/ui/Textarea';
import { Select } from '@/components/ui/Select';

const STATUS_OPTIONS = [
  { value: 'todo', label: 'To Do' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'done', label: 'Done' },
];

const PRIORITY_OPTIONS = [
  { value: 'low', label: 'Low' },
  { value: 'medium', label: 'Medium' },
  { value: 'high', label: 'High' },
];

export interface TaskFormData {
  title: string;
  description: string;
  priority: string;
  assignee_id: string;
  due_date: string;
  status: string;
}

interface TaskFormProps {
  form: TaskFormData;
  onChange: (form: TaskFormData) => void;
  onSubmit: (e: FormEvent) => void;
  onCancel: () => void;
  isEditing: boolean;
  saving: boolean;
  error: string;
  users: User[];
}

export function TaskForm({ form, onChange, onSubmit, onCancel, isEditing, saving, error, users }: TaskFormProps) {
  return (
    <form onSubmit={onSubmit} className="space-y-4">
      {error && (
        <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-md">
          {error}
        </div>
      )}

      <div className="space-y-2">
        <label htmlFor="task-title" className="text-sm font-medium">Title</label>
        <Input
          id="task-title"
          placeholder="Task title"
          value={form.title}
          onChange={(e) => onChange({ ...form, title: e.target.value })}
          autoFocus
        />
      </div>

      <div className="space-y-2">
        <label htmlFor="task-description" className="text-sm font-medium">Description (optional)</label>
        <Textarea
          id="task-description"
          placeholder="Describe the task"
          value={form.description}
          onChange={(e) => onChange({ ...form, description: e.target.value })}
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label htmlFor="task-priority" className="text-sm font-medium">Priority</label>
          <Select
            id="task-priority"
            value={form.priority}
            onChange={(e) => onChange({ ...form, priority: e.target.value })}
          >
            {PRIORITY_OPTIONS.map((p) => (
              <option key={p.value} value={p.value}>
                {p.label}
              </option>
            ))}
          </Select>
        </div>

        {isEditing && (
          <div className="space-y-2">
            <label htmlFor="task-status" className="text-sm font-medium">Status</label>
            <Select
              id="task-status"
              value={form.status}
              onChange={(e) => onChange({ ...form, status: e.target.value })}
            >
              {STATUS_OPTIONS.map((s) => (
                <option key={s.value} value={s.value}>
                  {s.label}
                </option>
              ))}
            </Select>
          </div>
        )}
      </div>

      <div className="space-y-2">
        <label htmlFor="task-assignee" className="text-sm font-medium">Assignee</label>
        <Select
          id="task-assignee"
          value={form.assignee_id}
          onChange={(e) => onChange({ ...form, assignee_id: e.target.value })}
        >
          <option value="">Unassigned</option>
          {users.map((u) => (
            <option key={u.id} value={u.id}>
              {u.name} ({u.email})
            </option>
          ))}
        </Select>
      </div>

      <div className="space-y-2">
        <label htmlFor="task-due-date" className="text-sm font-medium">Due Date (optional)</label>
        <Input
          id="task-due-date"
          type="date"
          value={form.due_date}
          onChange={(e) => onChange({ ...form, due_date: e.target.value })}
        />
      </div>

      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" disabled={saving}>
          {saving ? 'Saving...' : isEditing ? 'Update' : 'Create'}
        </Button>
      </div>
    </form>
  );
}
