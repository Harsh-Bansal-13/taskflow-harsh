import { useState, useMemo, useRef, useCallback, memo } from 'react';
import type { Task } from '@/types';
import { Card, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Calendar, Pencil, Trash2 } from 'lucide-react';
import { cn } from '@/lib/utils';

const COLUMNS = [
  { status: 'todo', label: 'To Do' },
  { status: 'in_progress', label: 'In Progress' },
  { status: 'done', label: 'Done' },
] as const;

const priorityBadgeVariant = (priority: string) => {
  switch (priority) {
    case 'high': return 'destructive' as const;
    case 'medium': return 'warning' as const;
    default: return 'outline' as const;
  }
};

interface KanbanBoardProps {
  tasks: Task[];
  currentUserId?: string;
  onStatusChange: (task: Task, status: string) => void;
  onEdit: (task: Task) => void;
  onDelete: (taskId: string) => void;
}

export const KanbanBoard = memo(function KanbanBoard({ tasks, currentUserId, onStatusChange, onEdit, onDelete }: KanbanBoardProps) {
  const [draggingId, setDraggingId] = useState<string | null>(null);
  const [overColumn, setOverColumn] = useState<string | null>(null);

  // Refs for touch drag — avoids stale closures inside touch event handlers
  const touchDragId = useRef<string | null>(null);
  const touchClone = useRef<HTMLElement | null>(null);

  const tasksByStatus = useMemo(() => {
    const map: Record<string, Task[]> = { todo: [], in_progress: [], done: [] };
    for (const t of tasks) {
      if (map[t.status]) map[t.status].push(t);
    }
    return map;
  }, [tasks]);

  // ── Mouse / pointer drag ──────────────────────────────────────────────────
  const handleDragStart = (e: React.DragEvent, taskId: string) => {
    e.dataTransfer.setData('text/plain', taskId);
    e.dataTransfer.effectAllowed = 'move';
    setDraggingId(taskId);
  };

  const handleDragEnd = () => {
    setDraggingId(null);
    setOverColumn(null);
  };

  const handleDragOver = (e: React.DragEvent, status: string) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    setOverColumn(status);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    if (!e.currentTarget.contains(e.relatedTarget as Node)) {
      setOverColumn(null);
    }
  };

  const handleDrop = (e: React.DragEvent, status: string) => {
    e.preventDefault();
    const taskId = e.dataTransfer.getData('text/plain');
    const task = tasks.find((t) => t.id === taskId);
    if (task && task.status !== status) {
      onStatusChange(task, status);
    }
    setDraggingId(null);
    setOverColumn(null);
  };

  // ── Touch drag ────────────────────────────────────────────────────────────
  /** Find the kanban column (data-column attribute) under a touch point */
  const columnAtPoint = (x: number, y: number): string | null => {
    // Temporarily hide the clone so elementFromPoint sees through it
    const clone = touchClone.current;
    if (clone) clone.style.pointerEvents = 'none';
    const el = document.elementFromPoint(x, y);
    if (clone) clone.style.pointerEvents = '';
    if (!el) return null;
    const col = (el as HTMLElement).closest('[data-column]');
    return col ? (col as HTMLElement).dataset.column ?? null : null;
  };

  const handleTouchStart = useCallback((e: React.TouchEvent, task: Task) => {
    const touch = e.touches[0];
    touchDragId.current = task.id;
    setDraggingId(task.id);

    // Build a visual clone that follows the finger
    const source = e.currentTarget as HTMLElement;
    const rect = source.getBoundingClientRect();
    const clone = source.cloneNode(true) as HTMLElement;
    clone.style.cssText = `
      position: fixed;
      left: ${rect.left}px;
      top: ${rect.top}px;
      width: ${rect.width}px;
      opacity: 0.85;
      pointer-events: none;
      z-index: 9999;
      transform: scale(1.04);
      transition: transform 0.1s;
    `;
    document.body.appendChild(clone);
    touchClone.current = clone;
  }, []);

  const handleTouchMove = useCallback((e: React.TouchEvent) => {
    if (!touchDragId.current) return;
    e.preventDefault(); // prevent scroll while dragging
    const touch = e.touches[0];

    // Move the clone
    const clone = touchClone.current;
    if (clone) {
      const rect = clone.getBoundingClientRect();
      clone.style.left = `${touch.clientX - rect.width / 2}px`;
      clone.style.top = `${touch.clientY - rect.height / 2}px`;
    }

    // Highlight the column under the finger
    const col = columnAtPoint(touch.clientX, touch.clientY);
    setOverColumn(col);
  }, []);

  const handleTouchEnd = useCallback((e: React.TouchEvent) => {
    if (!touchDragId.current) return;
    const touch = e.changedTouches[0];
    const col = columnAtPoint(touch.clientX, touch.clientY);

    if (col) {
      const task = tasks.find((t) => t.id === touchDragId.current);
      if (task && task.status !== col) {
        onStatusChange(task, col);
      }
    }

    // Cleanup
    touchClone.current?.remove();
    touchClone.current = null;
    touchDragId.current = null;
    setDraggingId(null);
    setOverColumn(null);
  }, [tasks, onStatusChange]);

  return (
    <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
      {COLUMNS.map((col) => {
        const colTasks = tasksByStatus[col.status] ?? [];
        const isOver = overColumn === col.status;

        return (
          <div
            key={col.status}
            data-column={col.status}
            className={cn(
              'rounded-lg border-2 transition-colors min-h-[200px] flex flex-col',
              isOver ? 'border-primary bg-primary/5' : 'border-border bg-muted/20'
            )}
            onDragOver={(e) => handleDragOver(e, col.status)}
            onDragLeave={handleDragLeave}
            onDrop={(e) => handleDrop(e, col.status)}
          >
            {/* Column header */}
            <div className="flex items-center justify-between px-3 py-2 border-b border-border">
              <span className="text-sm font-semibold">{col.label}</span>
              <span className="text-xs bg-muted text-muted-foreground rounded-full px-2 py-0.5">
                {colTasks.length}
              </span>
            </div>

            {/* Cards */}
            <div className="flex-1 p-2 space-y-2">
              {colTasks.map((task) => (
                <div
                  key={task.id}
                  draggable
                  onDragStart={(e) => handleDragStart(e, task.id)}
                  onDragEnd={handleDragEnd}
                  onDragOver={(e) => e.preventDefault()}
                  onTouchStart={(e) => handleTouchStart(e, task)}
                  onTouchMove={handleTouchMove}
                  onTouchEnd={handleTouchEnd}
                  className={cn(
                    'cursor-grab active:cursor-grabbing transition-opacity touch-none',
                    draggingId === task.id ? 'opacity-40' : 'opacity-100'
                  )}
                >
                  <Card className="hover:border-primary/40 transition-colors select-none">
                    <CardContent className="p-3">
                      <div className="flex items-start justify-between gap-2 mb-1">
                        <span className="text-sm font-medium leading-snug line-clamp-2">
                          {task.title}
                        </span>
                        <Badge variant={priorityBadgeVariant(task.priority)} className="shrink-0 text-xs">
                          {task.priority}
                        </Badge>
                      </div>

                      {task.description && (
                        <p className="text-xs text-muted-foreground line-clamp-2 mb-2">
                          {task.description}
                        </p>
                      )}

                      <div className="flex items-center gap-1.5 flex-wrap">
                        {task.assignee_id && (
                          <span className="text-xs bg-primary/10 text-primary px-1.5 py-0.5 rounded-full">
                            {task.assignee_id === currentUserId
                              ? 'Me'
                              : task.assignee_name || 'Assigned'}
                          </span>
                        )}
                        {task.due_date && (
                          <div className="flex items-center gap-0.5 text-xs text-muted-foreground">
                            <Calendar className="h-3 w-3" />
                            {new Date(task.due_date).toLocaleDateString()}
                          </div>
                        )}
                      </div>

                      <div className="flex justify-end gap-1 mt-2">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-7 w-7"
                          onClick={() => onEdit(task)}
                        >
                          <Pencil className="h-3 w-3" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-7 w-7"
                          onClick={() => onDelete(task.id)}
                        >
                          <Trash2 className="h-3 w-3 text-destructive" />
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                </div>
              ))}

              {colTasks.length === 0 && (
                <div
                  className={cn(
                    'flex items-center justify-center h-20 rounded-md border-2 border-dashed text-xs text-muted-foreground transition-colors',
                    isOver ? 'border-primary text-primary' : 'border-border'
                  )}
                >
                  {isOver ? 'Drop here' : 'No tasks'}
                </div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
});
KanbanBoard.displayName = 'KanbanBoard';
