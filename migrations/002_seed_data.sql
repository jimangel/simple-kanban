-- Seed data for initial setup

-- Create a default board
INSERT INTO boards (name, description)
VALUES ('Main Board', 'Default kanban board')
ON CONFLICT DO NOTHING;

-- Create default lists for the main board
INSERT INTO lists (board_id, name, position, color)
SELECT id, 'Backlog', 1.0, '#9ca3af'
FROM boards WHERE name = 'Main Board'
ON CONFLICT DO NOTHING;

INSERT INTO lists (board_id, name, position, color)
SELECT id, 'To Do', 2.0, '#60a5fa'
FROM boards WHERE name = 'Main Board'
ON CONFLICT DO NOTHING;

INSERT INTO lists (board_id, name, position, color)
SELECT id, 'In Progress', 3.0, '#fbbf24'
FROM boards WHERE name = 'Main Board'
ON CONFLICT DO NOTHING;

INSERT INTO lists (board_id, name, position, color)
SELECT id, 'Review', 4.0, '#c084fc'
FROM boards WHERE name = 'Main Board'
ON CONFLICT DO NOTHING;

INSERT INTO lists (board_id, name, position, color)
SELECT id, 'Done', 5.0, '#10b981'
FROM boards WHERE name = 'Main Board'
ON CONFLICT DO NOTHING;

-- Create some default labels
INSERT INTO labels (name, color) VALUES
    ('Bug', '#ef4444'),
    ('Feature', '#3b82f6'),
    ('Enhancement', '#8b5cf6'),
    ('Documentation', '#06b6d4'),
    ('High Priority', '#f97316'),
    ('Low Priority', '#6b7280')
ON CONFLICT DO NOTHING;