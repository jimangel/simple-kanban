package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/kanban-simple/internal/models"
)

// CardRepository handles database operations for cards
type CardRepository struct {
	db *sql.DB
}

// NewCardRepository creates a new card repository
func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

// Create creates a new card
func (r *CardRepository) Create(card *models.Card) error {
	// If position is not provided, calculate it
	if card.Position == 0 {
		var maxPosition sql.NullFloat64
		err := r.db.QueryRow(`
			SELECT MAX(position) FROM cards WHERE list_id = ?
		`, card.ListID).Scan(&maxPosition)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get max position: %w", err)
		}
		card.Position = maxPosition.Float64 + 1.0
	}

	query := `
		INSERT INTO cards (list_id, title, description, position, color, due_date, archived, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`
	now := time.Now()
	card.CreatedAt = now
	card.UpdatedAt = now

	err := r.db.QueryRow(
		query, card.ListID, card.Title, card.Description, card.Position,
		card.Color, card.DueDate, card.Archived, card.CreatedAt, card.UpdatedAt,
	).Scan(&card.ID)
	if err != nil {
		return fmt.Errorf("failed to create card: %w", err)
	}

	return nil
}

// GetByID retrieves a card by ID
func (r *CardRepository) GetByID(id int) (*models.Card, error) {
	card := &models.Card{}
	query := `
		SELECT id, list_id, title, description, position, color, due_date, archived, created_at, updated_at
		FROM cards
		WHERE id = ?
	`

	err := r.db.QueryRow(query, id).Scan(
		&card.ID, &card.ListID, &card.Title, &card.Description,
		&card.Position, &card.Color, &card.DueDate, &card.Archived,
		&card.CreatedAt, &card.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("card not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get card: %w", err)
	}

	return card, nil
}

// GetByListID retrieves all cards for a list
func (r *CardRepository) GetByListID(listID int, includeArchived bool) ([]models.Card, error) {
	query := `
		SELECT id, list_id, title, description, position, color, due_date, archived, created_at, updated_at
		FROM cards
		WHERE list_id = ?
	`
	args := []interface{}{listID}

	if !includeArchived {
		query += " AND archived = 0"
	}
	query += " ORDER BY position"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cards: %w", err)
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		err := rows.Scan(
			&card.ID, &card.ListID, &card.Title, &card.Description,
			&card.Position, &card.Color, &card.DueDate, &card.Archived,
			&card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		cards = append(cards, card)
	}

	// Ensure we never return nil, always return empty array
	if cards == nil {
		cards = []models.Card{}
	}

	return cards, nil
}

// Update updates a card
func (r *CardRepository) Update(card *models.Card) error {
	query := `
		UPDATE cards
		SET title = ?, description = ?, color = ?, due_date = ?, updated_at = ?
		WHERE id = ?
	`

	card.UpdatedAt = time.Now()
	result, err := r.db.Exec(
		query, card.Title, card.Description, card.Color,
		card.DueDate, card.UpdatedAt, card.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("card not found")
	}

	return nil
}

// Move moves a card to a different list and/or position
func (r *CardRepository) Move(cardID int, newListID int, newPosition float64) error {
	query := `
		UPDATE cards
		SET list_id = ?, position = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, newListID, newPosition, time.Now(), cardID)
	if err != nil {
		return fmt.Errorf("failed to move card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("card not found")
	}

	return nil
}

// Archive archives or unarchives a card
func (r *CardRepository) Archive(id int, archive bool) error {
	query := `
		UPDATE cards
		SET archived = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, archive, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to archive card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("card not found")
	}

	return nil
}

// Delete deletes a card
func (r *CardRepository) Delete(id int) error {
	query := `DELETE FROM cards WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("card not found")
	}

	return nil
}

// Search searches for cards based on criteria
func (r *CardRepository) Search(params models.SearchCardsRequest) ([]models.Card, error) {
	var conditions []string
	var args []interface{}

	query := `
		SELECT DISTINCT c.id, c.list_id, c.title, c.description, c.position,
		       c.color, c.due_date, c.archived, c.created_at, c.updated_at
		FROM cards c
		LEFT JOIN lists l ON c.list_id = l.id
		LEFT JOIN boards b ON l.board_id = b.id
		LEFT JOIN card_labels cl ON c.id = cl.card_id
		WHERE 1=1
	`

	// Add search conditions
	if params.Query != "" {
		conditions = append(conditions, "(c.title LIKE ? OR c.description LIKE ?)")
		searchTerm := "%" + params.Query + "%"
		args = append(args, searchTerm, searchTerm)
	}

	if params.BoardID != 0 {
		conditions = append(conditions, "b.id = ?")
		args = append(args, params.BoardID)
	}

	if params.ListID != 0 {
		conditions = append(conditions, "c.list_id = ?")
		args = append(args, params.ListID)
	}

	if params.Archived != nil {
		conditions = append(conditions, "c.archived = ?")
		args = append(args, *params.Archived)
	}

	if params.LabelID != 0 {
		conditions = append(conditions, "cl.label_id = ?")
		args = append(args, params.LabelID)
	}

	// Add conditions to query
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY c.created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search cards: %w", err)
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		err := rows.Scan(
			&card.ID, &card.ListID, &card.Title, &card.Description,
			&card.Position, &card.Color, &card.DueDate, &card.Archived,
			&card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		cards = append(cards, card)
	}

	// Ensure we never return nil, always return empty array
	if cards == nil {
		cards = []models.Card{}
	}

	return cards, nil
}

// GetAdjacentPositions finds positions for drag-drop reordering
func (r *CardRepository) GetAdjacentPositions(listID int, targetPosition float64) (float64, float64, error) {
	var prev, next sql.NullFloat64

	// Get the position before the target
	err := r.db.QueryRow(`
		SELECT MAX(position) FROM cards
		WHERE list_id = ? AND position < ?
	`, listID, targetPosition).Scan(&prev)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, fmt.Errorf("failed to get previous position: %w", err)
	}

	// Get the position after the target
	err = r.db.QueryRow(`
		SELECT MIN(position) FROM cards
		WHERE list_id = ? AND position > ?
	`, listID, targetPosition).Scan(&next)
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

// AddComment adds a comment to a card
func (r *CardRepository) AddComment(comment *models.Comment) error {
	query := `
		INSERT INTO comments (card_id, content, created_at)
		VALUES (?, ?, ?)
		RETURNING id
	`
	comment.CreatedAt = time.Now()

	err := r.db.QueryRow(query, comment.CardID, comment.Content, comment.CreatedAt).Scan(&comment.ID)
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	return nil
}

// GetComments retrieves all comments for a card
func (r *CardRepository) GetComments(cardID int) ([]models.Comment, error) {
	query := `
		SELECT id, card_id, content, created_at
		FROM comments
		WHERE card_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.CardID, &comment.Content, &comment.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}