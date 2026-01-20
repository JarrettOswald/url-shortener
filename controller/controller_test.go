package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockShortenerService struct {
	mock.Mock
}

func (m *MockShortenerService) ShortenURL(ctx context.Context, originalURL string) string {
	args := m.Called(ctx, originalURL)
	return args.String(0)
}

func (m *MockShortenerService) GetOriginalURL(ctx context.Context, shortKey string) (string, error) {
	args := m.Called(ctx, shortKey)
	return args.String(0), args.Error(1)
}

func setupRouter(c *Controller) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	c.RegisterRoutes(router)
	return router
}

func TestController_create_Success(t *testing.T) {
	mockService := new(MockShortenerService)
	controller := NewController(mockService)
	router := setupRouter(controller)

	requestBody := shortenRequest{
		URL: "https://example.com",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	mockService.On("ShortenURL", mock.Anything, "https://example.com").Return("abc123")

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response shortenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "abc123", response.URL)

	mockService.AssertExpectations(t)
}

func TestController_create_InvalidRequest(t *testing.T) {
	mockService := new(MockShortenerService)
	controller := NewController(mockService)
	router := setupRouter(controller)

	invalidBody := []byte(`{"invalid": "data"}`)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/", bytes.NewBuffer(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response errorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "failed to shorten url", response.Error)
}

func TestController_create_EmptyURL(t *testing.T) {
	mockService := new(MockShortenerService)
	controller := NewController(mockService)
	router := setupRouter(controller)

	requestBody := shortenRequest{
		URL: "",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response errorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "failed to shorten url", response.Error)
}

func TestController_create_MalformedJSON(t *testing.T) {
	mockService := new(MockShortenerService)
	controller := NewController(mockService)
	router := setupRouter(controller)

	invalidJSON := []byte(`{"url": "https://example.com"`)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestController_get_Success(t *testing.T) {
	mockService := new(MockShortenerService)
	controller := NewController(mockService)
	router := setupRouter(controller)

	mockService.On("GetOriginalURL", mock.Anything, "abc123").Return("https://example.com", nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/abc123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))

	mockService.AssertExpectations(t)
}

func TestController_get_NotFound(t *testing.T) {
	mockService := new(MockShortenerService)
	controller := NewController(mockService)
	router := setupRouter(controller)

	mockService.On("GetOriginalURL", mock.Anything, "notfound").Return("", errors.New("key not found"))

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notfound", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response errorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "url not found", response.Error)

	mockService.AssertExpectations(t)
}

func TestController_get_EmptyKey(t *testing.T) {
	mockService := new(MockShortenerService)
	controller := NewController(mockService)
	router := setupRouter(controller)

	mockService.On("GetOriginalURL", mock.Anything, "").Return("", errors.New("empty key"))

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestController_RegisterRoutes(t *testing.T) {
	mockService := new(MockShortenerService)
	controller := NewController(mockService)
	router := gin.New()

	controller.RegisterRoutes(router)

	routes := router.Routes()
	assert.Len(t, routes, 2)

	var hasPostRoute, hasGetRoute bool
	for _, route := range routes {
		if route.Method == "POST" && route.Path == "/api/v1/" {
			hasPostRoute = true
		}
		if route.Method == "GET" && route.Path == "/api/v1/:key" {
			hasGetRoute = true
		}
	}

	assert.True(t, hasPostRoute, "POST route should be registered")
	assert.True(t, hasGetRoute, "GET route should be registered")
}

func TestNewController(t *testing.T) {
	mockService := new(MockShortenerService)

	controller := NewController(mockService)

	assert.NotNil(t, controller)
	assert.Equal(t, mockService, controller.service)
}
