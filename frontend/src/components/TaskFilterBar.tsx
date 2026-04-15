import type { Task } from '@/types';
import { Button } from '@/components/ui/Button';
import { Select } from '@/components/ui/Select';
import {
  Plus,
  ListTodo,
  Clock,
  CheckCircle2,
  LayoutList,
  LayoutGrid,
} from 'lucide-react';

const STATUS_OPTIONS = [
  { value: 'todo', label: 'To Do', icon: ListTodo },
  { value: 'in_progress', label: 'In Progress', icon: Clock },
  { value: 'done', label: 'Done', icon: CheckCircle2 },
];

interface TaskFilterBarProps {
  tasks: Task[];
  statusFilter: string;
  onStatusFilterChange: (value: string) => void;
  assigneeFilter: string;
  onAssigneeFilterChange: (value: string) => void;
  assigneeIds: string[];
  assigneeMap: Map<string, string>;
  currentUserId?: string;
  view: 'list' | 'board';
  onViewChange: (view: 'list' | 'board') => void;
  onAddTask: () => void;
}

export function TaskFilterBar({
  tasks,
  statusFilter,
  onStatusFilterChange,
  assigneeFilter,
  onAssigneeFilterChange,
  assigneeIds,
  assigneeMap,
  currentUserId,
  view,
  onViewChange,
  onAddTask,
}: TaskFilterBarProps) {
  return (
    <div className="flex flex-col sm:flex-row gap-3 mb-6">
      <div className="flex gap-2 flex-wrap">
        <Button
          variant={statusFilter === '' ? 'default' : 'outline'}
          size="sm"
          onClick={() => onStatusFilterChange('')}
        >
          All ({tasks.length})
        </Button>
        {STATUS_OPTIONS.map((s) => {
          const count = tasks.filter((t) => t.status === s.value).length;
          return (
            <Button
              key={s.value}
              variant={statusFilter === s.value ? 'default' : 'outline'}
              size="sm"
              onClick={() => onStatusFilterChange(s.value)}
            >
              <s.icon className="h-3 w-3 mr-1" />
              {s.label} ({count})
            </Button>
          );
        })}
      </div>
      {assigneeIds.length > 0 && (
        <Select
          value={assigneeFilter}
          onChange={(e) => onAssigneeFilterChange(e.target.value)}
          className="h-9 w-44 text-sm"
        >
          <option value="">All Assignees</option>
          <option value="__unassigned__">Unassigned</option>
          {assigneeIds.map((aid) => (
            <option key={aid} value={aid}>
              {aid === currentUserId ? 'Me' : assigneeMap.get(aid) || 'Unknown'}
            </option>
          ))}
        </Select>
      )}
      <div className="sm:ml-auto flex items-center gap-2">
        <div className="flex rounded-md border overflow-hidden">
          <button
            type="button"
            className={`px-3 py-1.5 text-sm flex items-center gap-1 transition-colors ${
              view === 'list'
                ? 'bg-primary text-primary-foreground'
                : 'bg-background text-muted-foreground hover:bg-muted'
            }`}
            onClick={() => onViewChange('list')}
            aria-label="List view"
          >
            <LayoutList className="h-4 w-4" />
          </button>
          <button
            type="button"
            className={`px-3 py-1.5 text-sm flex items-center gap-1 transition-colors ${
              view === 'board'
                ? 'bg-primary text-primary-foreground'
                : 'bg-background text-muted-foreground hover:bg-muted'
            }`}
            onClick={() => onViewChange('board')}
            aria-label="Board view"
          >
            <LayoutGrid className="h-4 w-4" />
          </button>
        </div>
        <Button size="sm" onClick={onAddTask}>
          <Plus className="h-4 w-4 mr-1" /> Add Task
        </Button>
      </div>
    </div>
  );
}
