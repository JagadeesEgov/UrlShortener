package service

import (
	"net/http"
	"os"
	"strings"

	"urlShortner/models"
	"urlShortner/repository"
	"urlShortner/utils"

	"github.com/gin-gonic/gin"
)

type URLConverterService struct {
	repo *repository.PostgresRepository
}

func NewURLConverterService(repo *repository.PostgresRepository) *URLConverterService {
	return &URLConverterService{repo: repo}
}

func (s *URLConverterService) ShortenHandler(c *gin.Context) {
	var req models.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if !utils.ValidateURL(req.URL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid URL"})
		return
	}

	shortKey, err := s.repo.GetOrCreateShortKey(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	host := os.Getenv("HOST_NAME")
	contextPath := os.Getenv("SERVER_CONTEXT_PATH")
	// Ensure proper slashes
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
	if strings.HasPrefix(contextPath, "/") {
		contextPath = contextPath[1:]
	}
	if contextPath != "" && !strings.HasSuffix(contextPath, "/") {
		contextPath += "/"
	}
	shortURL := host + contextPath + shortKey
	c.JSON(http.StatusOK, gin.H{"shortUrl": shortURL})
}

func (s *URLConverterService) RedirectHandler(c *gin.Context) {
	shortKey := c.Param("key")
	longURL, err := s.repo.GetURL(shortKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Redirect(http.StatusMovedPermanently, longURL)
}
