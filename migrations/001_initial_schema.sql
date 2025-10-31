-- Initial schema for kanban board application

-- Boards table
CREATE TABLE IF NOT EXISTS boards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Lists/Columns table (using REAL for flexible positioning)
CREATE TABLE IF NOT EXISTS lists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    board_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    position REAL NOT NULL,
    color TEXT DEFAULT '#6b7280',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);

-- Cards/Tasks table
CREATE TABLE IF NOT EXISTS cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    list_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    position REAL NOT NULL,
    color TEXT,
    due_date DATETIME,
    archived INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (list_id) REFERENCES lists(id) ON DELETE CASCADE
);

-- Comments table for task updates
CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE CASCADE
);

-- Labels table for categorization
CREATE TABLE IF NOT EXISTS labels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Card labels junction table
CREATE TABLE IF NOT EXISTS card_labels (
    card_id INTEGER NOT NULL,
    label_id INTEGER NOT NULL,
    PRIMARY KEY (card_id, label_id),
    FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE CASCADE,
    FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_lists_board_position ON lists(board_id, position);
CREATE INDEX IF NOT EXISTS idx_cards_list_position ON cards(list_id, position);
CREATE INDEX IF NOT EXISTS idx_cards_archived ON cards(archived) WHERE archived = 0;
CREATE INDEX IF NOT EXISTS idx_comments_card ON comments(card_id);
CREATE INDEX IF NOT EXISTS idx_card_labels_card ON card_labels(card_id);
CREATE INDEX IF NOT EXISTS idx_card_labels_label ON card_labels(label_id);

-- Triggers to update timestamps
CREATE TRIGGER IF NOT EXISTS update_boards_timestamp
AFTER UPDATE ON boards
BEGIN
    UPDATE boards SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_lists_timestamp
AFTER UPDATE ON lists
BEGIN
    UPDATE lists SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_cards_timestamp
AFTER UPDATE ON cards
BEGIN
    UPDATE cards SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;