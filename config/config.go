package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	MaxConcurrentTasks int
	MaxFilesPerArchive int
	AllowedExtensions  []string
	DownloadDir        string
	ArchiveDir         string
	MaxDownloadSize    int64
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	maxTasksStr := os.Getenv("MAX_CONCURRENT_TASKS")
	maxConcurrentTasks, err := strconv.Atoi(maxTasksStr)
	if err != nil || maxConcurrentTasks <= 0 {
		maxConcurrentTasks = 3
	}

	maxFilesPerArchiveStr := os.Getenv("MAX_FILES_PER_ARCHIVE")
	maxFilesPerArchive, err := strconv.Atoi(maxFilesPerArchiveStr)
	if err != nil || maxFilesPerArchive <= 0 {
		maxFilesPerArchive = 3
	}

	allowedExt := os.Getenv("ALLOWED_EXTENSIONS")
	if allowedExt == "" {
		allowedExt = ".pdf,.jpeg"
	}
	allowedExtensions := strings.Split(allowedExt, ",")
	for i, ext := range allowedExtensions {
		ext = strings.TrimSpace(ext)
		ext = strings.ToLower(ext)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		allowedExtensions[i] = ext
	}
	//downloadDir служит в качестве временного хранилища для файлов, которые сервис скачивает
	//после архивации файлы удаляются
	downloadDir := os.Getenv("DOWNLOAD_DIR")
	if downloadDir == "" {
		downloadDir = "downloads"
	}
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		log.Fatalf("не удалось создать директорию для загрузок %s: %v", downloadDir, err)
	}

	archiveDir := os.Getenv("ARCHIVE_DIR")
	if archiveDir == "" {
		archiveDir = "archives"
	}
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		log.Fatalf("не удалось создать директорию для архивов %s: %v", archiveDir, err)
	}

	maxDownloadSizeMBStr := os.Getenv("MAX_DOWNLOAD_SIZE_MB")
	maxDownloadSizeMB, err := strconv.Atoi(maxDownloadSizeMBStr)
	if err != nil || maxDownloadSizeMB <= 0 {
		maxDownloadSizeMB = 50
	}

	return &Config{
		Port:               port,
		MaxConcurrentTasks: maxConcurrentTasks,
		MaxFilesPerArchive: maxFilesPerArchive,
		AllowedExtensions:  allowedExtensions,
		DownloadDir:        downloadDir,
		ArchiveDir:         archiveDir,
		MaxDownloadSize:    int64(maxDownloadSizeMB) * 1024 * 1024, // перевод MB → байты
	}
}
