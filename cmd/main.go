package main

import (
	"fmt"
	"log"
	"os"
	"urlShortner/internal/repository"
	"urlShortner/internal/service"
	migration "urlShortner/migrations"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
)

func main() {
	_ = godotenv.Load()
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery()) 

	// Always use Postgres repository
	pgRepo, err := repository.NewPostgresRepository()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Postgres repository connected");
	migration.RunMigrations(pgRepo.DB())
	svc := service.NewURLConverterService(pgRepo)

	r.POST("/shortener", svc.ShortenHandler)
	r.GET("/:key", svc.RedirectHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
