package http

import (
	"net/http"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.ShortenerService
}

func NewHandler(service service.ShortenerService) *Handler {
	return &Handler{service: service}
}

type shortenRequest struct {
	URL string `json:"url"`
}

func (h *Handler) Create(ctx *gin.Context) {
	var req shortenRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to shorten url"})
	}

	shortKey := h.service.ShortenURL(ctx, req.URL)

	ctx.JSON(http.StatusOK, gin.H{
		"url": shortKey,
	})
}

func (h *Handler) Get(ctx *gin.Context) {
	key := ctx.Param("key")

	originUrl, err := h.service.GetOriginalURL(ctx, key)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "url not found"})
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, originUrl)
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.POST("/", h.Create)
	router.GET("/:key", h.Get)
}
