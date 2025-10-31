package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kanban-simple/internal/api/middleware"
	"github.com/kanban-simple/internal/models"
	"github.com/kanban-simple/internal/repository"
)

// BoardHandler handles board-related HTTP requests
type BoardHandler struct {
	repo *repository.BoardRepository
}

// NewBoardHandler creates a new board handler
func NewBoardHandler(repo *repository.BoardRepository) *BoardHandler {
	return &BoardHandler{repo: repo}
}

// GetAll retrieves all boards
func (h *BoardHandler) GetAll(c *gin.Context) {
	boards, err := h.repo.GetAll()
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to retrieve boards")
		return
	}

	c.JSON(http.StatusOK, boards)
}

// GetByID retrieves a board by ID
func (h *BoardHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid board ID")
		return
	}

	board, err := h.repo.GetByID(id)
	if err != nil {
		if err.Error() == "board not found" {
			middleware.HandleError(c, http.StatusNotFound, "Board not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to retrieve board")
		}
		return
	}

	c.JSON(http.StatusOK, board)
}

// Create creates a new board
func (h *BoardHandler) Create(c *gin.Context) {
	var req models.CreateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	board := &models.Board{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.repo.Create(board); err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to create board")
		return
	}

	c.JSON(http.StatusCreated, board)
}

// Update updates a board
func (h *BoardHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid board ID")
		return
	}

	// Get existing board
	board, err := h.repo.GetByID(id)
	if err != nil {
		if err.Error() == "board not found" {
			middleware.HandleError(c, http.StatusNotFound, "Board not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to retrieve board")
		}
		return
	}

	// Bind update request
	var req models.UpdateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update fields if provided
	if req.Name != "" {
		board.Name = req.Name
	}
	if req.Description != "" {
		board.Description = req.Description
	}

	// Save updates
	if err := h.repo.Update(board); err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to update board")
		return
	}

	c.JSON(http.StatusOK, board)
}

// Delete deletes a board
func (h *BoardHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid board ID")
		return
	}

	if err := h.repo.Delete(id); err != nil {
		if err.Error() == "board not found" {
			middleware.HandleError(c, http.StatusNotFound, "Board not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to delete board")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Board deleted successfully"})
}