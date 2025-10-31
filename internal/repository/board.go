package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/kanban-simple/internal/models"
)

// BoardRepository handles database operations for boards
type BoardRepository struct {
	db *sql.DB
}

// NewBoardRepository creates a new board repository
func NewBoardRepository(db *sql.DB) *BoardRepository {
	return &BoardRepository{db: db}
}

// Create creates a new board
func (r *BoardRepository) Create(board *models.Board) error {
	query := `
		INSERT INTO boards (name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		RETURNING id
	`
	now := time.Now()
	board.CreatedAt = now
	board.UpdatedAt = now

	err := r.db.QueryRow(query, board.Name, board.Description, board.CreatedAt, board.UpdatedAt).Scan(&board.ID)
	if err != nil {
		return fmt.Errorf("failed to create board: %w", err)
	}

	return nil
}

// GetByID retrieves a board by ID
func (r *BoardRepository) GetByID(id int) (*models.Board, error) {
	board := &models.Board{}
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM boards
		WHERE id = ?
	`

	err := r.db.QueryRow(query, id).Scan(
		&board.ID, &board.Name, &board.Description,
		&board.CreatedAt, &board.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("board not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	return board, nil
}

// GetAll retrieves all boards
func (r *BoardRepository) GetAll() ([]models.Board, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM boards
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get boards: %w", err)
	}
	defer rows.Close()

	var boards []models.Board
	for rows.Next() {
		var board models.Board
		err := rows.Scan(
			&board.ID, &board.Name, &board.Description,
			&board.CreatedAt, &board.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan board: %w", err)
		}
		boards = append(boards, board)
	}

	return boards, nil
}

// Update updates a board
func (r *BoardRepository) Update(board *models.Board) error {
	query := `
		UPDATE boards
		SET name = ?, description = ?, updated_at = ?
		WHERE id = ?
	`

	board.UpdatedAt = time.Now()
	result, err := r.db.Exec(query, board.Name, board.Description, board.UpdatedAt, board.ID)
	if err != nil {
		return fmt.Errorf("failed to update board: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("board not found")
	}

	return nil
}

// GetByName retrieves a board by name
func (r *BoardRepository) GetByName(name string) (*models.Board, error) {
	board := &models.Board{}
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM boards
		WHERE name = ?
	`

	err := r.db.QueryRow(query, name).Scan(
		&board.ID, &board.Name, &board.Description,
		&board.CreatedAt, &board.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("board not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get board by name: %w", err)
	}

	return board, nil
}

// Delete deletes a board
func (r *BoardRepository) Delete(id int) error {
	query := `DELETE FROM boards WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete board: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("board not found")
	}

	return nil
}