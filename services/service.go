package services

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"return_zip_archive_service/config"
	"return_zip_archive_service/models"
	"return_zip_archive_service/utils"
)

var (
	ErrServerBusy            = errors.New("сервер занят, попробуйте позже")
	ErrTaskNotFound          = errors.New("задача не найдена")
	ErrMaxFilesReached       = errors.New("достигнуто максимальное количество файлов")
	ErrInvalidExtension      = errors.New("неподдерживаемый тип файла")
	ErrInvalidURL            = errors.New("некорректный URL")
	ErrTaskNotAcceptingFiles = errors.New("задача не принимает новые файлы (она уже обрабатывается или завершена)")
)

type TaskService struct {
	tasks          map[string]*models.Task
	taskMux        sync.RWMutex
	activeTasks    chan struct{}
	cfg            *config.Config
	fileDownloader FileDownloader
	archiver       Archiver
}

func NewTaskService(cfg *config.Config, fd FileDownloader, ar Archiver) *TaskService {
	return &TaskService{
		tasks:          make(map[string]*models.Task),
		activeTasks:    make(chan struct{}, cfg.MaxConcurrentTasks),
		cfg:            cfg,
		fileDownloader: fd,
		archiver:       ar,
	}
}
func (tm *TaskService) CreateTask() (string, error) {
	select {
	case tm.activeTasks <- struct{}{}:
		taskID := uuid.New().String()
		task := &models.Task{
			ID:        taskID,
			Status:    models.StatusWaiting,
			Files:     []models.FileInfo{},
			CreatedAt: time.Now(),
		}
		tm.taskMux.Lock()
		tm.tasks[taskID] = task
		tm.taskMux.Unlock()
		log.Printf("задача создана: %s", taskID)
		return taskID, nil
	default:
		return "", ErrServerBusy
	}
}

func (tm *TaskService) AddFileToTask(taskID string, fileURL string) error {
	tm.taskMux.RLock()
	task, exists := tm.tasks[taskID]
	tm.taskMux.RUnlock()

	if !exists {
		return ErrTaskNotFound
	}

	task.Mu.Lock()
	defer task.Mu.Unlock()

	if task.Status != models.StatusWaiting {
		return ErrTaskNotAcceptingFiles
	}

	if len(task.Files) >= tm.cfg.MaxFilesPerArchive {
		return ErrMaxFilesReached
	}

	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return ErrInvalidURL
	}

	ext := strings.ToLower(filepath.Ext(parsedURL.Path))
	if !utils.Contains(tm.cfg.AllowedExtensions, ext) {
		return ErrInvalidExtension
	}

	task.Files = append(task.Files, models.FileInfo{URL: fileURL, Downloaded: false})
	log.Printf("файл %s добавлен к задаче %s. Всего файлов: %d", fileURL, taskID, len(task.Files))

	if len(task.Files) == tm.cfg.MaxFilesPerArchive {
		task.Status = models.StatusRunning
		go tm.processTask(taskID)
	}

	return nil
}

func (tm *TaskService) GetTaskStatus(taskID string) (*models.Task, error) {
	tm.taskMux.RLock()
	task, exists := tm.tasks[taskID]
	tm.taskMux.RUnlock()

	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (tm *TaskService) processTask(taskID string) {
	defer func() {
		<-tm.activeTasks
		log.Printf("слот задачи освобожден для задачи %s", taskID)
	}()

	tm.taskMux.RLock()
	task := tm.tasks[taskID]
	tm.taskMux.RUnlock()

	if task == nil {
		log.Printf("ошибка: задача %s не найдена при обработке", taskID)
		return
	}

	task.Mu.Lock()
	if task.Status != models.StatusRunning {
		log.Printf("задача %s уже не в статусе 'processing' (%s). Пропуск повторной обработки.", taskID, task.Status)
		task.Mu.Unlock()
		return
	}
	defer task.Mu.Unlock()

	tempFiles := []string{}

	for i := range task.Files {
		fileInfo := &task.Files[i]
		log.Printf("скачивание файла: %s для задачи %s", fileInfo.URL, taskID)

		localFilePath, err := tm.fileDownloader.Download(fileInfo.URL, tm.cfg.DownloadDir)
		if err != nil {
			log.Printf("ошибка скачивания файла %s для задачи %s: %v", fileInfo.URL, taskID, err)

			task.Errors = append(task.Errors, models.FileError{
				URL:     fileInfo.URL,
				Message: "не удалось скачать файл",
				Detail:  err.Error(),
			})
			continue
		}

		fileInfo.LocalPath = localFilePath
		fileInfo.Downloaded = true
		tempFiles = append(tempFiles, localFilePath)
		log.Printf("файл %s успешно скачан в %s", fileInfo.URL, localFilePath)
	}

	if len(tempFiles) == 0 {
		task.Status = models.StatusError
		task.ErrorMessage = "не удалось скачать ни один из файлов или создать архив"
		task.CompletedAt = time.Now()
		log.Printf("задача %s завершена с ошибкой: %s", taskID, task.ErrorMessage)
		return
	}

	archiveFileName := fmt.Sprintf("%s.zip", taskID)
	archivePath := filepath.Join(tm.cfg.ArchiveDir, archiveFileName)
	archiveURL := fmt.Sprintf("/archives/%s", archiveFileName)

	if err := tm.archiver.CreateZip(tempFiles, archivePath); err != nil {
		task.Status = models.StatusError
		task.ErrorMessage = fmt.Sprintf("не удалось создать ZIP архив: %v", err)
		task.CompletedAt = time.Now()
		log.Printf("задача %s завершена с ошибкой: %s", taskID, task.ErrorMessage)
		return
	}

	task.Status = models.StatusSuccess
	task.ArchiveURL = archiveURL
	task.CompletedAt = time.Now()
	log.Printf("задача %s успешно завершена. Архив доступен по ссылке: %s", taskID, archiveURL)

	for _, f := range tempFiles {
		if err := os.Remove(f); err != nil {
			log.Printf("ошибка при удалении временного файла %s: %v", f, err)
		}
	}
	log.Printf("временные файлы для задачи %s удалены.", taskID)
}
