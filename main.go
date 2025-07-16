package main

import (
	"log"
	"net/http"
	"path/filepath"

	"return_zip_archive_service/config"
	"return_zip_archive_service/handlers"
	"return_zip_archive_service/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
	
	cfg := config.LoadConfig()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	fileDownloader := services.NewHTTPFileDownloader(cfg)
	zipArchiver := services.NewZipArchiver()

	taskService := services.NewTaskService(cfg, fileDownloader, zipArchiver)
	apiHandlers := handlers.NewHandlers(taskService, cfg)

	router.POST("/tasks", apiHandlers.CreateTaskHandler)
	router.POST("/tasks/:task_id/files", apiHandlers.AddFileHandler)
	router.GET("/tasks/:task_id/status", apiHandlers.GetTaskStatusHandler)

	router.StaticFS("/archives", http.Dir(filepath.Join(".", cfg.ArchiveDir)))

	log.Printf("Сервер запускается на порту %s", cfg.Port)
	log.Fatal(router.Run(":" + cfg.Port))
}
