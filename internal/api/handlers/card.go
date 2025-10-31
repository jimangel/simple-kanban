package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kanban-simple/internal/api/middleware"
	"github.com/kanban-simple/internal/models"
	"github.com/kanban-simple/internal/repository"
)

// CardHandler handles card-related HTTP requests
type CardHandler struct {
	cardRepo  *repository.CardRepository
	listRepo  *repository.ListRepository
	boardRepo *repository.BoardRepository
}

// NewCardHandler creates a new card handler
func NewCardHandler(cardRepo *repository.CardRepository, listRepo *repository.ListRepository, boardRepo *repository.BoardRepository) *CardHandler {
	return &CardHandler{
		cardRepo:  cardRepo,
		listRepo:  listRepo,
		boardRepo: boardRepo,
	}
}

// GetByID retrieves a card by ID
func (h *CardHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid card ID")
		return
	}

	card, err := h.cardRepo.GetByID(id)
	if err != nil {
		if err.Error() == "card not found" {
			middleware.HandleError(c, http.StatusNotFound, "Card not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to retrieve card")
		}
		return
	}

	// Get comments for the card
	comments, err := h.cardRepo.GetComments(id)
	if err == nil {
		card.Comments = comments
	}

	c.JSON(http.StatusOK, card)
}

// GetByListID retrieves all cards for a list
func (h *CardHandler) GetByListID(c *gin.Context) {
	listID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid list ID")
		return
	}

	// Check if archived cards should be included
	includeArchived := c.Query("archived") == "true"

	// Verify list exists
	if _, err := h.listRepo.GetByID(listID); err != nil {
		if err.Error() == "list not found" {
			middleware.HandleError(c, http.StatusNotFound, "List not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to verify list")
		}
		return
	}

	cards, err := h.cardRepo.GetByListID(listID, includeArchived)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to retrieve cards")
		return
	}

	c.JSON(http.StatusOK, cards)
}

// Create creates a new card
func (h *CardHandler) Create(c *gin.Context) {
	listID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid list ID")
		return
	}

	// Verify list exists
	if _, err := h.listRepo.GetByID(listID); err != nil {
		if err.Error() == "list not found" {
			middleware.HandleError(c, http.StatusNotFound, "List not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to verify list")
		}
		return
	}

	var req models.CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	card := &models.Card{
		ListID:      listID,
		Title:       req.Title,
		Description: req.Description,
		Position:    req.Position,
		Color:       req.Color,
		DueDate:     req.DueDate,
		Archived:    false,
	}

	if err := h.cardRepo.Create(card); err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to create card")
		return
	}

	c.JSON(http.StatusCreated, card)
}

// Update updates a card
func (h *CardHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid card ID")
		return
	}

	// Get existing card
	card, err := h.cardRepo.GetByID(id)
	if err != nil {
		if err.Error() == "card not found" {
			middleware.HandleError(c, http.StatusNotFound, "Card not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to retrieve card")
		}
		return
	}

	// Bind update request
	var req models.UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update fields if provided
	if req.Title != "" {
		card.Title = req.Title
	}
	if req.Description != "" {
		card.Description = req.Description
	}
	if req.Color != "" {
		card.Color = req.Color
	}
	if req.DueDate != nil {
		card.DueDate = req.DueDate
	}

	// Save updates
	if err := h.cardRepo.Update(card); err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to update card")
		return
	}

	c.JSON(http.StatusOK, card)
}

// Move moves a card to a different list and/or position
func (h *CardHandler) Move(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid card ID")
		return
	}

	var req models.MoveCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Verify card exists
	card, err := h.cardRepo.GetByID(id)
	if err != nil {
		if err.Error() == "card not found" {
			middleware.HandleError(c, http.StatusNotFound, "Card not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to retrieve card")
		}
		return
	}

	// Verify target list exists
	if _, err := h.listRepo.GetByID(req.ListID); err != nil {
		if err.Error() == "list not found" {
			middleware.HandleError(c, http.StatusNotFound, "Target list not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to verify target list")
		}
		return
	}

	// Move the card using the position calculated by the frontend
	if err := h.cardRepo.Move(id, req.ListID, req.Position); err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to move card")
		return
	}

	card.ListID = req.ListID
	card.Position = req.Position
	c.JSON(http.StatusOK, card)
}

// Archive archives a card
func (h *CardHandler) Archive(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid card ID")
		return
	}

	if err := h.cardRepo.Archive(id, true); err != nil {
		if err.Error() == "card not found" {
			middleware.HandleError(c, http.StatusNotFound, "Card not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to archive card")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card archived successfully"})
}

// Unarchive unarchives a card
func (h *CardHandler) Unarchive(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid card ID")
		return
	}

	if err := h.cardRepo.Archive(id, false); err != nil {
		if err.Error() == "card not found" {
			middleware.HandleError(c, http.StatusNotFound, "Card not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to unarchive card")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card unarchived successfully"})
}

// Delete deletes a card
func (h *CardHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid card ID")
		return
	}

	if err := h.cardRepo.Delete(id); err != nil {
		if err.Error() == "card not found" {
			middleware.HandleError(c, http.StatusNotFound, "Card not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to delete card")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card deleted successfully"})
}

// Search searches for cards based on criteria
func (h *CardHandler) Search(c *gin.Context) {
	var params models.SearchCardsRequest

	// Parse query parameters
	params.Query = c.Query("query")

	if boardID := c.Query("board_id"); boardID != "" {
		if id, err := strconv.Atoi(boardID); err == nil {
			params.BoardID = id
		}
	}

	if listID := c.Query("list_id"); listID != "" {
		if id, err := strconv.Atoi(listID); err == nil {
			params.ListID = id
		}
	}

	if archived := c.Query("archived"); archived != "" {
		archivedBool := archived == "true"
		params.Archived = &archivedBool
	}

	if labelID := c.Query("label_id"); labelID != "" {
		if id, err := strconv.Atoi(labelID); err == nil {
			params.LabelID = id
		}
	}

	cards, err := h.cardRepo.Search(params)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to search cards")
		return
	}

	c.JSON(http.StatusOK, cards)
}

// AddComment adds a comment to a card
func (h *CardHandler) AddComment(c *gin.Context) {
	cardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid card ID")
		return
	}

	// Verify card exists
	if _, err := h.cardRepo.GetByID(cardID); err != nil {
		if err.Error() == "card not found" {
			middleware.HandleError(c, http.StatusNotFound, "Card not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to verify card")
		}
		return
	}

	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	comment := &models.Comment{
		CardID:  cardID,
		Content: req.Content,
	}

	if err := h.cardRepo.AddComment(comment); err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to add comment")
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// GetComments retrieves all comments for a card
func (h *CardHandler) GetComments(c *gin.Context) {
	cardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid card ID")
		return
	}

	// Verify card exists
	if _, err := h.cardRepo.GetByID(cardID); err != nil {
		if err.Error() == "card not found" {
			middleware.HandleError(c, http.StatusNotFound, "Card not found")
		} else {
			middleware.HandleError(c, http.StatusInternalServerError, "Failed to verify card")
		}
		return
	}

	comments, err := h.cardRepo.GetComments(cardID)
	if err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to retrieve comments")
		return
	}

	c.JSON(http.StatusOK, comments)
}

// QuickCreate creates a card quickly (for bot integration)
func (h *CardHandler) QuickCreate(c *gin.Context) {
	type QuickCreateRequest struct {
		BoardName   string `json:"board_name,omitempty"`
		ListName    string `json:"list_name,omitempty"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description,omitempty"`
		Color       string `json:"color,omitempty"`
	}

	var req QuickCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Default to first board and "Backlog" list if not specified
	boardName := req.BoardName
	if boardName == "" {
		boardName = "Main Board"
	}

	listName := req.ListName
	if listName == "" {
		listName = "Backlog"
	}

	// Find the board by name
	board, err := h.boardRepo.GetByName(boardName)
	if err != nil {
		// If board not found, try to get the first board
		boards, err := h.boardRepo.GetAll()
		if err != nil || len(boards) == 0 {
			middleware.HandleError(c, http.StatusNotFound, "No boards available. Please create a board first.")
			return
		}
		board = &boards[0]
	}

	// Find the list by board ID and name
	list, err := h.listRepo.GetByBoardAndName(board.ID, listName)
	if err != nil {
		// If list not found, try to get the first list in the board
		lists, err := h.listRepo.GetByBoardID(board.ID)
		if err != nil || len(lists) == 0 {
			middleware.HandleError(c, http.StatusNotFound, "No lists available in the board. Please create a list first.")
			return
		}
		list = &lists[0]
	}

	// Create the card
	card := &models.Card{
		ListID:      list.ID,
		Title:       req.Title,
		Description: req.Description,
		Color:       req.Color,
		Archived:    false,
	}

	if err := h.cardRepo.Create(card); err != nil {
		middleware.HandleError(c, http.StatusInternalServerError, "Failed to create card")
		return
	}

	c.JSON(http.StatusCreated, card)
}