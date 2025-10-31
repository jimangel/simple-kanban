package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kanban-simple/internal/api/middleware"
	"github.com/kanban-simple/internal/models"
	"github.com/kanban-simple/internal/repository"
)

// ListHandler handles list-related HTTP requests
type ListHandler struct {
	listRepo  *repository.ListRepository
	boardRepo *repository.BoardRepository
}

// NewListHandler creates a new list handler
func NewListHandler(listRepo *repository.ListRepository, boardRepo *repository.BoardRepository) *ListHandler {
	return &ListHandler{
		listRepo:  listRepo,
		boardRepo: boardRepo,
	}
}

// GetByID retrieves a list by ID
func (h *ListHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid list ID")
		return
	}

	list, err := h.listRepo.GetByID(id)
	if err != nil {
		if err.Error() == "list not found" {
			middleware.HandleError(c, http.StatusNotFound, "List not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to retrieve list")
		}
		return
	}

	c.JSON(http.StatusOK, list)
}

// GetByBoardID retrieves all lists for a board
func (h *ListHandler) GetByBoardID(c *gin.Context) {
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid board ID")
		return
	}

	// Verify board exists
	if _, err := h.boardRepo.GetByID(boardID); err != nil {
		if err.Error() == "board not found" {
			middleware.HandleError(c, http.StatusNotFound, "Board not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to verify board")
		}
		return
	}

	lists, err := h.listRepo.GetByBoardID(boardID)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to retrieve lists")
		return
	}

	c.JSON(http.StatusOK, lists)
}

// Create creates a new list
func (h *ListHandler) Create(c *gin.Context) {
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid board ID")
		return
	}

	// Verify board exists
	if _, err := h.boardRepo.GetByID(boardID); err != nil {
		if err.Error() == "board not found" {
			middleware.HandleError(c, http.StatusNotFound, "Board not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to verify board")
		}
		return
	}

	var req models.CreateListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	list := &models.List{
		BoardID:  boardID,
		Name:     req.Name,
		Position: req.Position,
		Color:    req.Color,
	}

	// Set default color if not provided
	if list.Color == "" {
		list.Color = "#6b7280"
	}

	if err := h.listRepo.Create(list); err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to create list")
		return
	}

	c.JSON(http.StatusCreated, list)
}

// Update updates a list
func (h *ListHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid list ID")
		return
	}

	// Get existing list
	list, err := h.listRepo.GetByID(id)
	if err != nil {
		if err.Error() == "list not found" {
			middleware.HandleError(c, http.StatusNotFound, "List not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to retrieve list")
		}
		return
	}

	// Bind update request
	var req models.UpdateListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update fields if provided
	if req.Name != "" {
		list.Name = req.Name
	}
	if req.Position != 0 {
		list.Position = req.Position
	}
	if req.Color != "" {
		list.Color = req.Color
	}

	// Save updates
	if err := h.listRepo.Update(list); err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to update list")
		return
	}

	c.JSON(http.StatusOK, list)
}

// Move updates the position of a list
func (h *ListHandler) Move(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid list ID")
		return
	}

	var req models.MoveListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get the list to verify it exists
	list, err := h.listRepo.GetByID(id)
	if err != nil {
		if err.Error() == "list not found" {
			middleware.HandleError(c, http.StatusNotFound, "List not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to retrieve list")
		}
		return
	}

	// Calculate new position between adjacent lists
	prev, next, err := h.listRepo.GetAdjacentPositions(list.BoardID, req.Position)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to calculate position")
		return
	}

	// Calculate midpoint for new position
	newPosition := (prev + next) / 2

	// Update position
	if err := h.listRepo.UpdatePosition(id, newPosition); err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to move list")
		return
	}

	list.Position = newPosition
	c.JSON(http.StatusOK, list)
}

// Delete deletes a list
func (h *ListHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid list ID")
		return
	}

	if err := h.listRepo.Delete(id); err != nil {
		if err.Error() == "list not found" {
			middleware.HandleError(c, http.StatusNotFound, "List not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to delete list")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "List deleted successfully"})
}