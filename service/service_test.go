package service_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"urlShortner/models"
	"urlShortner/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) GetOrCreateShortKey(req models.ShortenRequest) (string, error) {
	args := m.Called(req)
	return args.String(0), args.Error(1)
}

func (m *MockRepo) GetURL(shortKey string) (string, error) {
	args := m.Called(shortKey)
	return args.String(0), args.Error(1)
}

func TestShortenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockRepo)
	service := service.NewURLConverterService(mockRepo)

	reqBody := `{"url":"https://test.com","validFrom":1000,"validTill":2000}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockRepo.On("GetOrCreateShortKey", mock.AnythingOfType("models.ShortenRequest")).Return("abc123", nil)

	service.ShortenHandler(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "shortUrl")
}

func TestShortenHandler_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockRepo)
	service := service.NewURLConverterService(mockRepo)

	reqBody := `{"invalid":"data"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	service.ShortenHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShortenHandler_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockRepo)
	service := service.NewURLConverterService(mockRepo)

	reqBody := `{"url":"https://test.com","validFrom":1000,"validTill":2000}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockRepo.On("GetOrCreateShortKey", mock.AnythingOfType("models.ShortenRequest")).Return("", errors.New("db error"))

	service.ShortenHandler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRedirectHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockRepo)
	service := service.NewURLConverterService(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "key", Value: "abc123"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/abc123", nil)

	mockRepo.On("GetURL", "abc123").Return("https://test.com", nil)

	service.RedirectHandler(c)
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://test.com", w.Header().Get("Location"))
}

func TestRedirectHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockRepo)
	service := service.NewURLConverterService(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "key", Value: "notfound"}}

	mockRepo.On("GetURL", "notfound").Return("", errors.New("not found"))

	service.RedirectHandler(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}