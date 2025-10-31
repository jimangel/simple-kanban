package models

import (
	"time"
)

// List represents a column in a kanban board
type List struct {
	ID        int       `json:"id" db:"id"`
	BoardID   int       `json:"board_id" db:"board_id"`
	Name      string    `json:"name" db:"name"`
	Position  float64   `json:"position" db:"position"`
	Color     string    `json:"color" db:"color"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Cards     []Card    `json:"cards,omitempty"` // Populated when needed
}

// CreateListRequest represents the request to create a new list
type CreateListRequest struct {
	Name     string  `json:"name" binding:"required,min=1,max=255"`
	Position float64 `json:"position,omitempty"`
	Color    string  `json:"color,omitempty"`
}

// UpdateListRequest represents the request to update a list
type UpdateListRequest struct {
	Name     string  `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Position float64 `json:"position,omitempty"`
	Color    string  `json:"color,omitempty"`
}

// MoveListRequest represents the request to move a list
type MoveListRequest struct {
	Position float64 `json:"position" binding:"required"`
}