package models

import (
	"time"

	"gorm.io/gorm"
)

type Checklist struct {
	gorm.Model
	Title       string     `json:"title"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at"`
	TaskId      uint       `json:"task_id"`
	UserId      uint       `json:"user_id"`
}
