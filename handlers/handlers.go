package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"urlShortner/config"
	"urlShortner/models"
	"urlShortner/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler holds the HTTP handlers for the URL shortening service
type Handler struct {
	urlService *service.URLService
	config     *config.Config
	log        *logrus.Logger
}

// NewHandler creates a new handler instance
func NewHandler(urlService *service.URLService, cfg *config.Config, log *logrus.Logger) *Handler {
	return &Handler{
		urlService: urlService,
		config:     cfg,
		log:        log,
	}
}

// ShortenURL handles URL shortening requests
func (h *Handler) ShortenURL(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var request models.ShortenRequest
	
	// Bind JSON request body
	if err := c.ShouldBindJSON(&request); err != nil {
		h.log.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"INVALID_REQUEST",
			"Invalid request format: "+err.Error(),
		))
		return
	}

	// Additional validation
	if err := request.Validate(); err != nil {
		h.log.WithError(err).WithField("request", request.String()).Warn("Request validation failed")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VALIDATION_ERROR",
			err.Error(),
		))
		return
	}

	// Log headers for debugging (matching Java implementation)
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[strings.ToLower(key)] = values[0]
		}
	}
	h.log.WithFields(logrus.Fields{
		"headers": headers,
		"request": request.String(),
	}).Debug("Processing shorten URL request")

	// Determine tenant ID based on multi-instance configuration
	var tenantID string
	if !h.config.App.IsMultiInstance {
		tenantID = h.config.App.StateLevelTenantID
	} else {
		// Extract state-specific tenant ID from ULB level tenant
		tenantIDHeader := c.GetHeader("tenantid")
		if tenantIDHeader == "" {
			h.log.Error("TenantId header is missing in multi-instance mode")
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"INVALID_TENANTID",
				"TenantId not present in header",
			))
			return
		}

		// Extract state level tenant from ULB tenant
		tenantID = h.urlService.ExtractStateLevelTenant(tenantIDHeader)
		h.log.WithFields(logrus.Fields{
			"ulbTenantId":   tenantIDHeader,
			"stateTenantId": tenantID,
		}).Debug("Extracted state level tenant ID")
	}

	// Validate URL using service
	if !h.urlService.ValidateURL(request.URL) {
		h.log.WithField("url", request.URL).Warn("URL validation failed")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"URL_SHORTENING_INVALID_URL",
			"Please enter a valid URL",
		))
		return
	}

	// Shorten URL
	shortenedURL, err := h.urlService.ShortenURL(ctx, &request, tenantID, h.config.App.IsMultiInstance)
	if err != nil {
		h.log.WithError(err).WithFields(logrus.Fields{
			"url":      request.URL,
			"tenantId": tenantID,
		}).Error("Failed to shorten URL")
		
		// Determine error type and response
		if strings.Contains(err.Error(), "invalid URL") {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"URL_SHORTENING_INVALID_URL",
				err.Error(),
			))
		} else if strings.Contains(err.Error(), "tenant") {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"EG_TENANT_HOST_NOT_FOUND_ERR",
				err.Error(),
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				"URL_SHORTENING_FAILED",
				"Failed to shorten URL",
			))
		}
		return
	}

	h.log.WithFields(logrus.Fields{
		"originalUrl":  request.URL,
		"shortenedUrl": shortenedURL,
		"tenantId":     tenantID,
	}).Info("URL shortened successfully")

	// Return shortened URL as plain text (matching Java implementation)
	c.String(http.StatusOK, shortenedURL)
}

// RedirectURL handles URL redirection requests
func (h *Handler) RedirectURL(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	id := c.Param("id")
	
	if id == "" {
		h.log.Error("ID parameter is missing")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"INVALID_REQUEST",
			"ID parameter is required",
		))
		return
	}

	// Sanitize ID parameter
	id = strings.TrimSpace(id)
	if id == "" {
		h.log.Error("ID parameter is empty after trimming")
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"INVALID_REQUEST",
			"ID parameter cannot be empty",
		))
		return
	}

	h.log.WithField("id", id).Debug("Processing redirect request")

	// Get original URL
	longURL, err := h.urlService.GetLongURLFromID(ctx, id)
	if err != nil {
		h.log.WithError(err).WithField("id", id).Warn("Failed to get long URL")
		
		// Determine error type
		if strings.Contains(err.Error(), "does not exist") {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"INVALID_REQUEST",
				"Invalid Key",
			))
		} else if strings.Contains(err.Error(), "expired") {
			c.JSON(http.StatusGone, models.NewErrorResponse(
				"URL_EXPIRED",
				"URL has expired",
			))
		} else if strings.Contains(err.Error(), "not yet active") {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"URL_NOT_ACTIVE",
				"URL is not yet active",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				"INTERNAL_ERROR",
				"Failed to retrieve URL",
			))
		}
		return
	}

	h.log.WithFields(logrus.Fields{
		"id":      id,
		"longURL": longURL,
	}).Info("Redirecting to original URL")

	// Redirect to original URL
	c.Redirect(http.StatusFound, longURL)
}

// GetURLDetails handles URL details retrieval requests
func (h *Handler) GetURLDetails(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	id := c.Param("id")
	
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"INVALID_REQUEST",
			"ID parameter is required",
		))
		return
	}

	// Get URL details
	details, err := h.urlService.GetURLDetails(ctx, id)
	if err != nil {
		h.log.WithError(err).WithField("id", id).Error("Failed to get URL details")
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			"INVALID_REQUEST",
			"URL not found",
		))
		return
	}

	c.JSON(http.StatusOK, details)
}

// DeleteURL handles URL deletion requests
func (h *Handler) DeleteURL(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	id := c.Param("id")
	
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"INVALID_REQUEST",
			"ID parameter is required",
		))
		return
	}

	// Delete URL
	err := h.urlService.DeleteURL(ctx, id)
	if err != nil {
		h.log.WithError(err).WithField("id", id).Error("Failed to delete URL")
		
		if strings.Contains(err.Error(), "does not exist") {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"INVALID_REQUEST",
				"URL not found",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				"DELETE_FAILED",
				"Failed to delete URL",
			))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "URL deleted successfully",
		"id":      id,
	})
}

// HealthCheck provides a health check endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	response := models.NewHealthCheckResponse("egov-url-shortening-go", "1.0.0")

	// Check service health
	if err := h.urlService.HealthCheck(ctx); err != nil {
		h.log.WithError(err).Error("Health check failed")
		response.Status = "DOWN"
		response.AddCheck("service", "DOWN: "+err.Error())
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	response.AddCheck("service", "UP")
	response.AddCheck("database", "UP")
	response.AddCheck("repository", "UP")

	c.JSON(http.StatusOK, response)
}

// GetStats provides service statistics
func (h *Handler) GetStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	stats := h.urlService.GetServiceStats(ctx)
	c.JSON(http.StatusOK, stats)
}

// CleanupExpiredURLs triggers cleanup of expired URLs
func (h *Handler) CleanupExpiredURLs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	deletedCount, err := h.urlService.CleanupExpiredURLs(ctx)
	if err != nil {
		h.log.WithError(err).Error("Failed to cleanup expired URLs")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"CLEANUP_FAILED",
			"Failed to cleanup expired URLs",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Cleanup completed",
		"deletedCount": deletedCount,
	})
}

// Middleware for request logging
func (h *Handler) RequestLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		h.log.WithFields(logrus.Fields{
			"status":     param.StatusCode,
			"method":     param.Method,
			"path":       param.Path,
			"ip":         param.ClientIP,
			"user_agent": param.Request.UserAgent(),
			"latency":    param.Latency.String(),
		}).Info("HTTP Request")
		return ""
	})
}

// Middleware for error handling
func (h *Handler) ErrorHandlingMiddleware() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, recovered interface{}) {
		h.log.WithField("panic", recovered).Error("Panic recovered")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"INTERNAL_ERROR",
			"Internal server error",
		))
	})
}

// Middleware for request timeout
func (h *Handler) TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
