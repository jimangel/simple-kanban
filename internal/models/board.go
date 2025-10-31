package models

import (
	"time"
)

// Board represents a kanban board
type Board struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Lists       []List    `json:"lists,omitempty"` // Populated when needed
}

// CreateBoardRequest represents the request to create a new board
type CreateBoardRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
}

// UpdateBoardRequest represents the request to update a board
type UpdateBoardRequest struct {
	Name        string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description string `json:"description,omitempty"`
}