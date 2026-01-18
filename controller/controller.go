package controller

import (
	"net/http"
	"url-shortener/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service service.ShortenerService
}

func NewController(service service.ShortenerService) *Controller {
	return &Controller{service: service}
}

type shortenRequest struct {
	URL string `json:"url" binding:"required" example:"https://example.com"`
}

type shortenResponse struct {
	URL string `json:"url" example:"abc123"`
}

type errorResponse struct {
	Error string `json:"error" example:"url not found"`
}

// create godoc
//
//	@Summary		Shorten URL
//	@Description	create a shortened URL from a long URL
//	@Tags			urls
//	@Accept			json
//	@Produce		json
//	@Param			request	body		shortenRequest	true	"URL to shorten"
//	@Success		200		{object}	shortenResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/ [post]
func (c *Controller) create(ctx *gin.Context) {
	var req shortenRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "failed to shorten url"})
	}

	shortKey := c.service.ShortenURL(ctx, req.URL)

	ctx.JSON(http.StatusOK, shortenResponse{
		URL: shortKey,
	})
}

// get godoc
//
//	@Summary		get original URL
//	@Description	Redirect to the original URL by short key
//	@Tags			urls
//	@Produce		json
//	@Param			key	path		string	true	"Short URL key"
//	@Success		301	{string}	string	"Redirect to original URL"
//	@Failure		404	{object}	errorResponse
//	@Router			/api/v1/{key} [get]
func (c *Controller) get(ctx *gin.Context) {
	key := ctx.Param("key")

	originUrl, err := c.service.GetOriginalURL(ctx, key)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "url not found"})
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, originUrl)
}

func (c *Controller) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.POST("/", c.create)
		api.GET("/:key", c.get)
	}
}
