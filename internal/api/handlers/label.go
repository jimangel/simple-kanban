package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kanban-simple/internal/models"
	"github.com/kanban-simple/internal/repository"
)

// LabelHandler handles label-related HTTP requests
type LabelHandler struct {
	labelRepo *repository.LabelRepository
	cardRepo  *repository.CardRepository
}

// NewLabelHandler creates a new label handler
func NewLabelHandler(labelRepo *repository.LabelRepository, cardRepo *repository.CardRepository) *LabelHandler {
	return &LabelHandler{
		labelRepo: labelRepo,
		cardRepo:  cardRepo,
	}
}

// GetAll retrieves all labels
func (h *LabelHandler) GetAll(c *gin.Context) {
	labels, err := h.labelRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, labels)
}

// GetByID retrieves a specific label
func (h *LabelHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid label ID",
		})
		return
	}

	label, err := h.labelRepo.GetByID(id)
	if err != nil {
		if err.Error() == "label not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Label not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, label)
}

// Create creates a new label
func (h *LabelHandler) Create(c *gin.Context) {
	var req models.CreateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	label, err := h.labelRepo.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, label)
}

// Update updates a label
func (h *LabelHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid label ID",
		})
		return
	}

	var req models.CreateLabelRequest // Reusing the same request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	label, err := h.labelRepo.Update(id, req.Name, req.Color)
	if err != nil {
		if err.Error() == "label not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Label not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, label)
}

// Delete deletes a label
func (h *LabelHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid label ID",
		})
		return
	}

	if err := h.labelRepo.Delete(id); err != nil {
		if err.Error() == "label not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Label not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// AssignToCard assigns a label to a card
func (h *LabelHandler) AssignToCard(c *gin.Context) {
	cardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid card ID",
		})
		return
	}

	labelID, err := strconv.Atoi(c.Param("label_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid label ID",
		})
		return
	}

	// Verify card exists
	_, err = h.cardRepo.GetByID(cardID)
	if err != nil {
		if err.Error() == "card not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Card not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	// Verify label exists
	_, err = h.labelRepo.GetByID(labelID)
	if err != nil {
		if err.Error() == "label not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Label not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	// Assign label to card
	if err := h.labelRepo.AssignToCard(cardID, labelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Label assigned to card successfully",
	})
}

// RemoveFromCard removes a label from a card
func (h *LabelHandler) RemoveFromCard(c *gin.Context) {
	cardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid card ID",
		})
		return
	}

	labelID, err := strconv.Atoi(c.Param("label_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid label ID",
		})
		return
	}

	if err := h.labelRepo.RemoveFromCard(cardID, labelID); err != nil {
		if err.Error() == "label assignment not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Label assignment not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetCardLabels gets all labels for a card
func (h *LabelHandler) GetCardLabels(c *gin.Context) {
	cardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid card ID",
		})
		return
	}

	// Verify card exists
	_, err = h.cardRepo.GetByID(cardID)
	if err != nil {
		if err.Error() == "card not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Card not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	labels, err := h.labelRepo.GetCardLabels(cardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, labels)
}