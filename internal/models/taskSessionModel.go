package models

import (
	"time"

	"gorm.io/gorm"
)

type TaskSession struct {
	gorm.Model
	ID        uint       `json:"id" gorm:"primaryKey"`
	TaskID    uint       `json:"task_id" gorm:"not null"`
	Task      Task       `json:"task" gorm:"foreignKey:TaskID"`
	UserID    uint       `json:"user_id" gorm:"not null"`
	User      User       `json:"user" gorm:"foreignKey:UserID"`
	StartTime time.Time  `json:"start_time" gorm:"not null"`
	EndTime   *time.Time `json:"end_time"`
	Duration  int64      `json:"duration"`
}
