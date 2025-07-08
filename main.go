package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"urlShortner/config"
	"urlShortner/handlers"
	"urlShortner/repository"
	"urlShortner/service"
	"urlShortner/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.Info("Starting eGov URL Shortening Service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger.WithFields(logrus.Fields{
		"port":             cfg.Server.Port,
		"database_enabled": cfg.Database.Enabled,
		"multi_instance":   cfg.App.IsMultiInstance,
	}).Info("Configuration loaded")

	// Initialize HashID converter
	hashConverter, err := utils.NewHashIDConverter(&cfg.HashIDs)
	if err != nil {
		log.Fatalf("Failed to initialize HashID converter: %v", err)
	}

	// Initialize URL validator
	urlValidator := utils.NewURLValidator()

	// Initialize repository based on configuration
	var urlRepo repository.URLRepository
	if cfg.Database.Enabled {
		// Use PostgreSQL
		urlRepo, err = repository.NewPostgresRepository(&cfg.Database, logger)
		if err != nil {
			log.Fatalf("Failed to initialize PostgreSQL repository: %v", err)
		}
		logger.Info("Using PostgreSQL repository")
	} else {
		// Use Redis
		urlRepo = repository.NewRedisRepository(&cfg.Redis, logger)
		logger.Info("Using Redis repository")
	}

	// Initialize service
	urlService := service.NewURLService(urlRepo, hashConverter, urlValidator, cfg, logger)

	// Initialize handlers
	handler := handlers.NewHandler(urlService, cfg, logger)

	// Setup Gin router
	if cfg.Server.Port == 8091 { // Production mode
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()
	
	// Use custom middleware
	router.Use(handler.RequestLoggingMiddleware())
	router.Use(handler.ErrorHandlingMiddleware())
	router.Use(handler.TimeoutMiddleware(60 * time.Second))

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, tenantid, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Setup routes
	setupRoutes(router, handler, cfg)

	// Start server
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.WithField("port", cfg.Server.Port).Info("Starting HTTP server")

	// Graceful shutdown
	go func() {
		if err := router.Run(port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	
	// Close repository connections
	if closeableRepo, ok := urlRepo.(interface{ Close() error }); ok {
		if err := closeableRepo.Close(); err != nil {
			logger.WithError(err).Error("Error closing repository connection")
		}
	}
	
	logger.Info("Server shutdown complete")
}

// setupRoutes configures the HTTP routes
func setupRoutes(router *gin.Engine, handler *handlers.Handler, cfg *config.Config) {
	// Health check endpoint (without context path)
	router.GET("/health", handler.HealthCheck)
	
	// Service stats endpoint
	router.GET("/stats", handler.GetStats)

	// API routes with context path
	api := router.Group(cfg.Server.ContextPath)
	{
		// URL shortening endpoint
		api.POST("/shortener", handler.ShortenURL)
		
		// URL redirection endpoint
		api.GET("/:id", handler.RedirectURL)
		
		// URL details endpoint
		api.GET("/details/:id", handler.GetURLDetails)
		
		// URL deletion endpoint
		api.DELETE("/:id", handler.DeleteURL)
		
		// Admin endpoints
		admin := api.Group("/admin")
		{
			// Cleanup expired URLs
			admin.POST("/cleanup", handler.CleanupExpiredURLs)
			
			// Service statistics
			admin.GET("/stats", handler.GetStats)
		}
	}

	// Also add routes without context path for direct access
	router.POST("/shortener", handler.ShortenURL)
	router.GET("/:id", handler.RedirectURL)
}
