package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"urlShortner/internal/models"
	"urlShortner/internal/repository"
	"urlShortner/internal/utils"

	"github.com/gin-gonic/gin"
)

type URLConverterService struct {
	repo repository.URLRepository
}

func NewURLConverterService(repo repository.URLRepository) *URLConverterService {
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

	//  Read flags
	multiInstance := os.Getenv("MULTI_INSTANCE") == "true"
	stateTenantID := os.Getenv("STATE_LEVEL_TENANT_ID")
	hostName := os.Getenv("HOST_NAME")
	hostMapJson := os.Getenv("EGOV_UI_APP_HOST_MAP")

	tenantID := c.GetHeader("tenantid")
	if multiInstance {
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "TenantId not present in header"})
			return
		}
		// Simulate getStateLevelTenant(ulbTenantID)
		parts := strings.Split(tenantID, ".")
		if len(parts) < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant format"})
			return
		}
		stateCode := parts[0]

		// Parse the host map
		var hostMap map[string]string
		if err := json.Unmarshal([]byte(hostMapJson), &hostMap); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid host map config"})
			return
		}

		stateHost, ok := hostMap[stateCode]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Hostname not found for tenantId: %s", stateCode)})
			return
		}
		hostName = stateHost
	} else {
		tenantID = stateTenantID // fallback
	}

	// Normalize host and context path
	hostName = strings.TrimRight(hostName, "/")
	contextPath := strings.Trim(os.Getenv("SERVER_CONTEXT_PATH"), "/")

	// Short key generation 
	shortKey, err := s.repo.GetOrCreateShortKey(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	shortURL := fmt.Sprintf("%s/%s/%s", hostName, contextPath, shortKey)
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
