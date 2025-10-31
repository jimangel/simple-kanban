package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kanban-simple/internal/api/handlers"
	"github.com/kanban-simple/internal/api/middleware"
	"github.com/kanban-simple/internal/repository"
)

// Repositories holds all repository instances
type Repositories struct {
	Board *repository.BoardRepository
	List  *repository.ListRepository
	Card  *repository.CardRepository
	Label *repository.LabelRepository
}

// NewRouter creates and configures the Gin router
func NewRouter(repos *Repositories) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	router.Use(middleware.ErrorHandler())

	// Initialize handlers
	boardHandler := handlers.NewBoardHandler(repos.Board)
	listHandler := handlers.NewListHandler(repos.List, repos.Board)
	cardHandler := handlers.NewCardHandler(repos.Card, repos.List, repos.Board)
	labelHandler := handlers.NewLabelHandler(repos.Label, repos.Card)

	// API routes
	api := router.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})

		// Board endpoints
		boards := api.Group("/boards")
		{
			boards.GET("", boardHandler.GetAll)
			boards.POST("", boardHandler.Create)
			boards.GET("/:id", boardHandler.GetByID)
			boards.PUT("/:id", boardHandler.Update)
			boards.DELETE("/:id", boardHandler.Delete)

			// Lists endpoints (nested under boards)
			boards.GET("/:id/lists", listHandler.GetByBoardID)
			boards.POST("/:id/lists", listHandler.Create)
		}

		// List endpoints
		lists := api.Group("/lists")
		{
			lists.GET("/:id", listHandler.GetByID)
			lists.PUT("/:id", listHandler.Update)
			lists.PATCH("/:id/move", listHandler.Move)
			lists.DELETE("/:id", listHandler.Delete)

			// Cards endpoints (nested under lists)
			lists.GET("/:id/cards", cardHandler.GetByListID)
			lists.POST("/:id/cards", cardHandler.Create)
		}

		// Card endpoints
		cards := api.Group("/cards")
		{
			cards.GET("", cardHandler.Search)
			cards.GET("/:id", cardHandler.GetByID)
			cards.PUT("/:id", cardHandler.Update)
			cards.PATCH("/:id/move", cardHandler.Move)
			cards.POST("/:id/archive", cardHandler.Archive)
			cards.POST("/:id/unarchive", cardHandler.Unarchive)
			cards.DELETE("/:id", cardHandler.Delete)

			// Comments
			cards.GET("/:id/comments", cardHandler.GetComments)
			cards.POST("/:id/comments", cardHandler.AddComment)
		}

		// Quick card creation for bots
		api.POST("/cards/quick", cardHandler.QuickCreate)

		// Search endpoint
		api.GET("/search", cardHandler.Search)

		// Label endpoints
		labels := api.Group("/labels")
		{
			labels.GET("", labelHandler.GetAll)
			labels.POST("", labelHandler.Create)
			labels.GET("/:id", labelHandler.GetByID)
			labels.PUT("/:id", labelHandler.Update)
			labels.DELETE("/:id", labelHandler.Delete)
		}

		// Card-Label associations
		api.POST("/cards/:id/labels/:label_id", labelHandler.AssignToCard)
		api.DELETE("/cards/:id/labels/:label_id", labelHandler.RemoveFromCard)
		api.GET("/cards/:id/labels", labelHandler.GetCardLabels)
	}

	// Serve OpenAPI specification
	router.StaticFile("/openapi.yaml", "./openapi.yaml")

	// Static files (web UI)
	router.Static("/static", "./web/static")
	router.GET("/", func(c *gin.Context) {
		c.File("./web/static/index.html")
	})

	return router
}