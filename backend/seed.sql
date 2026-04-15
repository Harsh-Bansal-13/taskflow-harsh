-- seed.sql
-- Password: password123  (bcrypt cost 12)
-- Pre-generated hash so the seed is deterministic and doesn't need a runtime bcrypt call.
INSERT INTO users (id, name, email, password)
VALUES (
    'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
    'Test User',
    'test@example.com',
    '$2a$12$hzIkb/h2DeWTYs7GtEzaw.gdJ6R/E7S7Hmq8coia.Flrf2otceb2a'
) ON CONFLICT (email) DO NOTHING;

INSERT INTO projects (id, name, description, owner_id)
VALUES (
    'b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e',
    'Website Redesign',
    'Complete overhaul of the company website for Q2 launch',
    'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d'
) ON CONFLICT DO NOTHING;

INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, created_by, due_date)
VALUES
(
    'c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f',
    'Design homepage mockup',
    'Create wireframes and high-fidelity mockups for the new homepage',
    'in_progress',
    'high',
    'b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e',
    'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
    'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
    '2026-04-20'
),
(
    'd4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f80',
    'Set up CI/CD pipeline',
    'Configure GitHub Actions for automated testing and deployment',
    'todo',
    'medium',
    'b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e',
    NULL,
    'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
    '2026-04-25'
),
(
    'e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8091',
    'Write API documentation',
    'Document all REST endpoints with request/response examples',
    'done',
    'low',
    'b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e',
    'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
    'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
    '2026-04-15'
)
ON CONFLICT DO NOTHING;
