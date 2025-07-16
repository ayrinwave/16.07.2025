package handlers

import (
	"errors"
	"log"
	"net/http"
	"return_zip_archive_service/config"
	"return_zip_archive_service/models"
	"return_zip_archive_service/services"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Message string `json:"message"`
}
type Handlers struct {
	TaskService *services.TaskService
	Config      *config.Config
}

func NewHandlers(tm *services.TaskService, cfg *config.Config) *Handlers {
	return &Handlers{
		TaskService: tm,
		Config:      cfg,
	}
}
func (h *Handlers) CreateTaskHandler(c *gin.Context) {
	taskID, err := h.TaskService.CreateTask()
	if err != nil {
		if errors.Is(err, services.ErrServerBusy) {
			c.JSON(http.StatusTooManyRequests, APIError{Message: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, APIError{Message: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task_id": taskID})
	log.Printf("создана новая задача с ID: %s", taskID)
}

func (h *Handlers) AddFileHandler(c *gin.Context) {
	taskID := c.Param("task_id")

	var req struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIError{Message: "некорректное тело запроса"})
		return
	}

	if req.URL == "" {
		c.JSON(http.StatusBadRequest, APIError{Message: "URL пустой"})
		return
	}

	err := h.TaskService.AddFileToTask(taskID, req.URL)
	if err != nil {
		log.Printf("ошибка при добавлении файла к задаче %s: %v", taskID, err)
		if errors.Is(err, services.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, APIError{Message: err.Error()})
		} else if errors.Is(err, services.ErrInvalidExtension) ||
			errors.Is(err, services.ErrMaxFilesReached) ||
			errors.Is(err, services.ErrInvalidURL) ||
			errors.Is(err, services.ErrTaskNotAcceptingFiles) {
			c.JSON(http.StatusBadRequest, APIError{Message: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, APIError{Message: "ошибка сервера: " + err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "файл успешно добавлен"})
	log.Printf("файл добавлен к задаче %s", taskID)
}

func (h *Handlers) GetTaskStatusHandler(c *gin.Context) {
	taskID := c.Param("task_id")

	task, err := h.TaskService.GetTaskStatus(taskID)
	if err != nil {
		log.Printf("ошибка при получении статуса задачи %s: %v", taskID, err)
		c.JSON(http.StatusNotFound, APIError{Message: err.Error()})
		return
	}

	resp := models.TaskStatusResponse{
		ID:           task.ID,
		Status:       task.Status,
		ArchiveURL:   task.ArchiveURL,
		ErrorMessage: task.ErrorMessage,
		FileErrors:   task.Errors,
	}

	c.JSON(http.StatusOK, resp)
	log.Printf("статус задачи %s, статус: %s", taskID, task.Status)
}
