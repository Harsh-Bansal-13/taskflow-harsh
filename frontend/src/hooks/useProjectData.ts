import { useState, useEffect, useCallback } from 'react';
import { getProject, getUsers } from '@/api/endpoints';
import type { ProjectWithTasks, User } from '@/types';

export function useProjectData(id: string | undefined) {
  const [project, setProject] = useState<ProjectWithTasks | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchProject = useCallback(async () => {
    if (!id) return;
    try {
      setLoading(true);
      const res = await getProject(id);
      setProject(res.data);
    } catch {
      setError('Failed to load project');
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchProject();
    getUsers(id).then((res) => setUsers(res.data)).catch(() => {});
  }, [fetchProject, id]);

  // Real-time SSE subscription: refetch when a task mutation is broadcast
  useEffect(() => {
    if (!id) return;
    const token = localStorage.getItem('token');
    if (!token) return;
    const base = import.meta.env.VITE_API_URL || 'http://localhost:8080';
    const es = new EventSource(`${base}/projects/${id}/events?token=${encodeURIComponent(token)}`);
    es.onmessage = () => { fetchProject(); };
    es.onerror = () => { es.close(); };
    return () => { es.close(); };
  }, [id, fetchProject]);

  return { project, setProject, users, loading, error, fetchProject };
}
