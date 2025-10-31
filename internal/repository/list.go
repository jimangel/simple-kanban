package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/kanban-simple/internal/models"
)

// ListRepository handles database operations for lists
type ListRepository struct {
	db *sql.DB
}

// NewListRepository creates a new list repository
func NewListRepository(db *sql.DB) *ListRepository {
	return &ListRepository{db: db}
}

// Create creates a new list
func (r *ListRepository) Create(list *models.List) error {
	// If position is not provided, calculate it
	if list.Position == 0 {
		var maxPosition sql.NullFloat64
		err := r.db.QueryRow(`
			SELECT MAX(position) FROM lists WHERE board_id = ?
		`, list.BoardID).Scan(&maxPosition)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get max position: %w", err)
		}
		list.Position = maxPosition.Float64 + 1.0
	}

	query := `
		INSERT INTO lists (board_id, name, position, color, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id
	`
	now := time.Now()
	list.CreatedAt = now
	list.UpdatedAt = now

	err := r.db.QueryRow(
		query, list.BoardID, list.Name, list.Position,
		list.Color, list.CreatedAt, list.UpdatedAt,
	).Scan(&list.ID)
	if err != nil {
		return fmt.Errorf("failed to create list: %w", err)
	}

	return nil
}

// GetByID retrieves a list by ID
func (r *ListRepository) GetByID(id int) (*models.List, error) {
	list := &models.List{}
	query := `
		SELECT id, board_id, name, position, color, created_at, updated_at
		FROM lists
		WHERE id = ?
	`

	err := r.db.QueryRow(query, id).Scan(
		&list.ID, &list.BoardID, &list.Name, &list.Position,
		&list.Color, &list.CreatedAt, &list.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("list not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get list: %w", err)
	}

	return list, nil
}

// GetByBoardID retrieves all lists for a board
func (r *ListRepository) GetByBoardID(boardID int) ([]models.List, error) {
	query := `
		SELECT id, board_id, name, position, color, created_at, updated_at
		FROM lists
		WHERE board_id = ?
		ORDER BY position
	`

	rows, err := r.db.Query(query, boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lists: %w", err)
	}
	defer rows.Close()

	var lists []models.List
	for rows.Next() {
		var list models.List
		err := rows.Scan(
			&list.ID, &list.BoardID, &list.Name, &list.Position,
			&list.Color, &list.CreatedAt, &list.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan list: %w", err)
		}
		lists = append(lists, list)
	}

	return lists, nil
}

// Update updates a list
func (r *ListRepository) Update(list *models.List) error {
	query := `
		UPDATE lists
		SET name = ?, position = ?, color = ?, updated_at = ?
		WHERE id = ?
	`

	list.UpdatedAt = time.Now()
	result, err := r.db.Exec(
		query, list.Name, list.Position, list.Color,
		list.UpdatedAt, list.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update list: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("list not found")
	}

	return nil
}

// UpdatePosition updates only the position of a list
func (r *ListRepository) UpdatePosition(id int, position float64) error {
	query := `
		UPDATE lists
		SET position = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, position, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update list position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("list not found")
	}

	return nil
}

// GetByBoardAndName retrieves a list by board ID and list name
func (r *ListRepository) GetByBoardAndName(boardID int, name string) (*models.List, error) {
	list := &models.List{}
	query := `
		SELECT id, board_id, name, position, color, created_at, updated_at
		FROM lists
		WHERE board_id = ? AND name = ?
	`

	err := r.db.QueryRow(query, boardID, name).Scan(
		&list.ID, &list.BoardID, &list.Name, &list.Position,
		&list.Color, &list.CreatedAt, &list.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("list not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get list by board and name: %w", err)
	}

	return list, nil
}

// Delete deletes a list
func (r *ListRepository) Delete(id int) error {
	query := `DELETE FROM lists WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete list: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("list not found")
	}

	return nil
}

// GetAdjacentPositions finds positions for drag-drop reordering
func (r *ListRepository) GetAdjacentPositions(boardID int, targetPosition float64) (float64, float64, error) {
	var prev, next sql.NullFloat64

	// Get the position before the target
	err := r.db.QueryRow(`
		SELECT MAX(position) FROM lists
		WHERE board_id = ? AND position < ?
	`, boardID, targetPosition).Scan(&prev)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, fmt.Errorf("failed to get previous position: %w", err)
	}

	// Get the position after the target
	err = r.db.QueryRow(`
		SELECT MIN(position) FROM lists
		WHERE board_id = ? AND position > ?
	`, boardID, targetPosition).Scan(&next)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, fmt.Errorf("failed to get next position: %w", err)
	}

	if !prev.Valid {
		prev.Float64 = 0
	}
	if !next.Valid {
		next.Float64 = prev.Float64 + 2
	}

	return prev.Float64, next.Float64, nil
}