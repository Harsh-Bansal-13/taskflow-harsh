-- 000002_add_created_by_index.up.sql
CREATE INDEX idx_tasks_created_by ON tasks(created_by);
