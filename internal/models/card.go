package models

import (
	"time"
)

// Card represents a task/ticket in a kanban list
type Card struct {
	ID          int       `json:"id" db:"id"`
	ListID      int       `json:"list_id" db:"list_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description,omitempty" db:"description"`
	Position    float64   `json:"position" db:"position"`
	Color       string    `json:"color,omitempty" db:"color"`
	DueDate     *time.Time `json:"due_date,omitempty" db:"due_date"`
	Archived    bool      `json:"archived" db:"archived"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Comments    []Comment `json:"comments,omitempty"` // Populated when needed
	Labels      []Label   `json:"labels,omitempty"`   // Populated when needed
}

// Comment represents a comment on a card
type Comment struct {
	ID        int       `json:"id" db:"id"`
	CardID    int       `json:"card_id" db:"card_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Label represents a label for categorization
type Label struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Color     string    `json:"color" db:"color"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CreateCardRequest represents the request to create a new card
type CreateCardRequest struct {
	Title       string     `json:"title" binding:"required,min=1,max=255"`
	Description string     `json:"description,omitempty"`
	Position    float64    `json:"position,omitempty"`
	Color       string     `json:"color,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// UpdateCardRequest represents the request to update a card
type UpdateCardRequest struct {
	Title       string     `json:"title,omitempty" binding:"omitempty,min=1,max=255"`
	Description string     `json:"description,omitempty"`
	Color       string     `json:"color,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// MoveCardRequest represents the request to move a card
type MoveCardRequest struct {
	ListID   int     `json:"list_id" binding:"required"`
	Position float64 `json:"position" binding:"required"`
}

// CreateCommentRequest represents the request to create a comment
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1"`
}

// CreateLabelRequest represents the request to create a label
type CreateLabelRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=50"`
	Color string `json:"color" binding:"required"`
}

// SearchCardsRequest represents card search parameters
type SearchCardsRequest struct {
	Query    string `json:"query,omitempty" form:"query"`
	BoardID  int    `json:"board_id,omitempty" form:"board_id"`
	ListID   int    `json:"list_id,omitempty" form:"list_id"`
	Archived *bool  `json:"archived,omitempty" form:"archived"`
	LabelID  int    `json:"label_id,omitempty" form:"label_id"`
}