package models

import (
	"time"

	"gorm.io/gorm"
)

type Checklist struct {
	gorm.Model
	ID          uint       `json:"id" gorm:"primaryKey"`
	Title       string     `json:"title"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at"`
	TaskId      uint       `json:"task_id"`
	UserId      uint       `json:"user_id"`
}
