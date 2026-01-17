package main

import (
	"context"
	"time"
	"url-shortener/controller"
	_ "url-shortener/docs"
	"url-shortener/repository"
	"url-shortener/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title			Swagger Example API
//	@version		1.0
//	@description	This is a sample server url-shortner server.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
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
	h := controller.NewController(svc)

	router := gin.Default()
	h.RegisterRoutes(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":8080")
}
