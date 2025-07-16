package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"return_zip_archive_service/config"
	"return_zip_archive_service/utils"
)

type FileDownloader interface {
	Download(fileURL, downloadDir string) (string, error)
}
type HTTPFileDownloader struct {
	config *config.Config
}

func NewHTTPFileDownloader(cfg *config.Config) *HTTPFileDownloader {
	return &HTTPFileDownloader{
		config: cfg,
	}
}

func (d *HTTPFileDownloader) Download(fileURL, downloadDir string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	// user agent нужен для того чтобы сервисы не блокировали и не обрывали скачивание
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MyDownloader/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка при запросе URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("не удалось скачать файл, статус: %s", resp.Status)
	}

	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("не удалось разобрать URL: %w", err)
	}

	fileName := filepath.Base(parsedURL.Path)
	if fileName == "" || fileName == "/" {
		fileName = uuid.New().String() + ".tmp"
	}

	uniqueFileName := utils.GetUniqueFileName(downloadDir, fileName)
	filePath := filepath.Join(downloadDir, uniqueFileName)

	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("не удалось создать файл %s: %w", filePath, err)
	}
	defer out.Close()

	limitedReader := io.LimitReader(resp.Body, 50*1024*1024)
	_, err = io.Copy(out, limitedReader)
	if err != nil {
		return "", fmt.Errorf("ошибка при записи файла: %w", err)
	}

	return filePath, nil
}
