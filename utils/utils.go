package utils

import (
	"os"
	"path/filepath"
	"strconv"
)

func Contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
func GetUniqueFileName(dir, filename string) string {
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]
	fullPath := filepath.Join(dir, filename)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return filename
	}

	for i := 1; ; i++ {
		newFileName := name + "(" + strconv.Itoa(i) + ")" + ext
		newFullPath := filepath.Join(dir, newFileName)
		if _, err := os.Stat(newFullPath); os.IsNotExist(err) {
			return newFileName
		}
	}
}
