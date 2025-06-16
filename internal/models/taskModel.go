package models

import "gorm.io/gorm"

// type Status string

// const (
// 	Completed  Status = "completed"
// 	Incomplete Status = "incomplete"
// )

type Task struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	UserId      uint   `json:"user_id"`
	TimeSpend   uint   `json:"time_spend"`
	Streak      uint   `json:"streak"`
}
