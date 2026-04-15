import { useMemo, useState } from 'react';
import type { Task, User } from '@/types';

export function useTaskFilters(tasks: Task[], users: User[]) {
  const [statusFilter, setStatusFilter] = useState('');
  const [assigneeFilter, setAssigneeFilter] = useState('');

  // Build assignee lookup from users + task data
  const assigneeMap = useMemo(() => {
    const map = new Map<string, string>();
    users.forEach((u) => map.set(u.id, u.name));
    tasks.forEach((t) => {
      if (t.assignee_id && t.assignee_name && !map.has(t.assignee_id)) {
        map.set(t.assignee_id, t.assignee_name);
      }
    });
    return map;
  }, [users, tasks]);

  // Derive unique assignees for filter
  const assigneeIds = useMemo(
    () =>
      Array.from(
        new Set(tasks.filter((t) => t.assignee_id).map((t) => t.assignee_id!))
      ),
    [tasks]
  );

  // Filter tasks
  const filteredTasks = useMemo(
    () =>
      tasks.filter((t) => {
        if (statusFilter && t.status !== statusFilter) return false;
        if (assigneeFilter === '__unassigned__' && t.assignee_id) return false;
        if (assigneeFilter && assigneeFilter !== '__unassigned__' && t.assignee_id !== assigneeFilter) return false;
        return true;
      }),
    [tasks, statusFilter, assigneeFilter]
  );

  return {
    statusFilter,
    setStatusFilter,
    assigneeFilter,
    setAssigneeFilter,
    assigneeMap,
    assigneeIds,
    filteredTasks,
  };
}
