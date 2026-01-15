package main

import (
	"context"
	"time"
	handler "url-shortener/internal/http"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := repository.Config{
		Addr:        "localhost:6379",
		Password:    "test1234",
		User:        "default",
		DB:          0,
		MaxRetries:  5,
		DialTimeout: 10 * time.Second,
		Timeout:     5 * time.Second,
	}

	rdb, _ := repository.NewClient(context.Background(), cfg)

	repo := repository.NewRedisRepository(rdb)
	svc := service.NewShortenerService(repo)
	h := handler.NewHandler(svc)

	router := gin.Default()
	h.RegisterRoutes(router)

	router.Run(":8080")
}
