package repository

import (
	"database/sql"
	"fmt"

	"github.com/kanban-simple/internal/models"
)

// LabelRepository handles label database operations
type LabelRepository struct {
	db *sql.DB
}

// NewLabelRepository creates a new label repository
func NewLabelRepository(db *sql.DB) *LabelRepository {
	return &LabelRepository{db: db}
}

// Create creates a new label
func (r *LabelRepository) Create(req *models.CreateLabelRequest) (*models.Label, error) {
	query := `
		INSERT INTO labels (name, color)
		VALUES (?, ?)
		RETURNING id, name, color, created_at`

	var label models.Label
	err := r.db.QueryRow(query, req.Name, req.Color).Scan(
		&label.ID,
		&label.Name,
		&label.Color,
		&label.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create label: %w", err)
	}

	return &label, nil
}

// GetAll retrieves all labels
func (r *LabelRepository) GetAll() ([]models.Label, error) {
	query := `
		SELECT id, name, color, created_at
		FROM labels
		ORDER BY name ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}
	defer rows.Close()

	var labels []models.Label
	for rows.Next() {
		var label models.Label
		err := rows.Scan(
			&label.ID,
			&label.Name,
			&label.Color,
			&label.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labels = append(labels, label)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating labels: %w", err)
	}

	if labels == nil {
		labels = []models.Label{}
	}

	return labels, nil
}

// GetByID retrieves a label by ID
func (r *LabelRepository) GetByID(id int) (*models.Label, error) {
	query := `
		SELECT id, name, color, created_at
		FROM labels
		WHERE id = ?`

	var label models.Label
	err := r.db.QueryRow(query, id).Scan(
		&label.ID,
		&label.Name,
		&label.Color,
		&label.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("label not found")
		}
		return nil, fmt.Errorf("failed to get label: %w", err)
	}

	return &label, nil
}

// Update updates a label
func (r *LabelRepository) Update(id int, name, color string) (*models.Label, error) {
	query := `
		UPDATE labels
		SET name = ?, color = ?
		WHERE id = ?
		RETURNING id, name, color, created_at`

	var label models.Label
	err := r.db.QueryRow(query, name, color, id).Scan(
		&label.ID,
		&label.Name,
		&label.Color,
		&label.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("label not found")
		}
		return nil, fmt.Errorf("failed to update label: %w", err)
	}

	return &label, nil
}

// Delete deletes a label
func (r *LabelRepository) Delete(id int) error {
	// First, remove all associations with cards
	_, err := r.db.Exec("DELETE FROM card_labels WHERE label_id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to remove label associations: %w", err)
	}

	// Then delete the label
	result, err := r.db.Exec("DELETE FROM labels WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("label not found")
	}

	return nil
}

// AssignToCard assigns a label to a card
func (r *LabelRepository) AssignToCard(cardID, labelID int) error {
	// Check if assignment already exists
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM card_labels WHERE card_id = ? AND label_id = ?)",
		cardID, labelID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check existing assignment: %w", err)
	}

	if exists {
		return nil // Already assigned, no need to do anything
	}

	// Create the assignment
	_, err = r.db.Exec(
		"INSERT INTO card_labels (card_id, label_id) VALUES (?, ?)",
		cardID, labelID,
	)
	if err != nil {
		return fmt.Errorf("failed to assign label to card: %w", err)
	}

	return nil
}

// RemoveFromCard removes a label from a card
func (r *LabelRepository) RemoveFromCard(cardID, labelID int) error {
	result, err := r.db.Exec(
		"DELETE FROM card_labels WHERE card_id = ? AND label_id = ?",
		cardID, labelID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove label from card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("label assignment not found")
	}

	return nil
}

// GetCardLabels gets all labels for a card
func (r *LabelRepository) GetCardLabels(cardID int) ([]models.Label, error) {
	query := `
		SELECT l.id, l.name, l.color, l.created_at
		FROM labels l
		INNER JOIN card_labels cl ON l.id = cl.label_id
		WHERE cl.card_id = ?
		ORDER BY l.name ASC`

	rows, err := r.db.Query(query, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card labels: %w", err)
	}
	defer rows.Close()

	var labels []models.Label
	for rows.Next() {
		var label models.Label
		err := rows.Scan(
			&label.ID,
			&label.Name,
			&label.Color,
			&label.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labels = append(labels, label)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating labels: %w", err)
	}

	if labels == nil {
		labels = []models.Label{}
	}

	return labels, nil
}