package models

import (
	"sync"
	"time"
)

type TaskStatus string

const (
	StatusWaiting TaskStatus = "waiting"
	StatusRunning TaskStatus = "running"
	StatusSuccess TaskStatus = "success"
	StatusError   TaskStatus = "error"
)

type FileInfo struct {
	URL        string `json:"url"`
	LocalPath  string
	Downloaded bool
	Error      string
}
type Task struct {
	ID           string     `json:"id"`
	Status       TaskStatus `json:"status"`
	Files        []FileInfo `json:"files"`
	ArchiveURL   string     `json:"archive_url,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  time.Time  `json:"completed_at,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	Mu           sync.Mutex
	Errors       []FileError `json:"file_errors,omitempty"`
}
type TaskStatusResponse struct {
	ID           string      `json:"id"`
	Status       TaskStatus  `json:"status"`
	ArchiveURL   string      `json:"archive_url,omitempty"`
	ErrorMessage string      `json:"error_message,omitempty"`
	FileErrors   []FileError `json:"file_errors,omitempty"`
}
type FileError struct {
	URL     string `json:"url"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}
