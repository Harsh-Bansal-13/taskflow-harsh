import { memo } from 'react';
import type { Task } from '@/types';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Select } from '@/components/ui/Select';
import { Calendar, Pencil, Trash2 } from 'lucide-react';

const STATUS_OPTIONS = [
  { value: 'todo', label: 'To Do' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'done', label: 'Done' },
];

const priorityBadgeVariant = (priority: string) => {
  switch (priority) {
    case 'high': return 'destructive' as const;
    case 'medium': return 'warning' as const;
    default: return 'outline' as const;
  }
};

interface TaskCardProps {
  task: Task;
  currentUserId?: string;
  onStatusChange: (task: Task, status: string) => void;
  onEdit: (task: Task) => void;
  onDelete: (taskId: string) => void;
}

export const TaskCard = memo(function TaskCard({ task, currentUserId, onStatusChange, onEdit, onDelete }: TaskCardProps) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center gap-3">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap mb-1">
          <h3 className="font-medium truncate">{task.title}</h3>
          <Badge variant={priorityBadgeVariant(task.priority)}>
            {task.priority}
          </Badge>
        </div>
        {task.description && (
          <p className="text-sm text-muted-foreground line-clamp-2">
            {task.description}
          </p>
        )}
        <div className="flex items-center gap-2 mt-1 flex-wrap">
          {task.assignee_id && (
            <span className="text-xs bg-primary/10 text-primary px-2 py-0.5 rounded-full">
              {task.assignee_id === currentUserId
                ? 'Assigned to me'
                : task.assignee_name || `${task.assignee_id.slice(0, 8)}...`}
            </span>
          )}
          {task.due_date && (
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <Calendar className="h-3 w-3" />
              Due {new Date(task.due_date).toLocaleDateString()}
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center gap-2 flex-shrink-0">
        <Select
          value={task.status}
          onChange={(e) => onStatusChange(task, e.target.value)}
          className="h-9 w-36 text-xs"
        >
          {STATUS_OPTIONS.map((s) => (
            <option key={s.value} value={s.value}>
              {s.label}
            </option>
          ))}
        </Select>

        <Button variant="ghost" size="icon" onClick={() => onEdit(task)}>
          <Pencil className="h-4 w-4" />
        </Button>
        <Button variant="ghost" size="icon" onClick={() => onDelete(task.id)}>
          <Trash2 className="h-4 w-4 text-destructive" />
        </Button>
      </div>
    </div>
  );
});
