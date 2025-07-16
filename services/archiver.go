package services

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Archiver interface {
	CreateZip(files []string, archivePath string) error
}
type ZipArchiver struct{}

func NewZipArchiver() *ZipArchiver {
	return &ZipArchiver{}
}

func (za *ZipArchiver) CreateZip(files []string, archivePath string) error {
	newZipFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("не удалось создать ZIP файл: %v", err)
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	if len(files) == 0 {
		return errors.New("нет файлов для архивации")
	}

	for _, file := range files {
		err := addFileToZip(zipWriter, file)
		if err != nil {
			return fmt.Errorf("не удалось добавить файл %s в архив: %v", file, err)
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл %s: %v", filename, err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("не удалось получить информацию о файле %s: %v", filename, err)
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		file.Close()
		return fmt.Errorf("не удалось создать заголовок файла для %s: %v", filename, err)
	}

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		file.Close()
		return fmt.Errorf("не удалось создать заголовок в ZIP: %v", err)
	}

	_, err = io.Copy(writer, file)
	file.Close()
	return err
}
