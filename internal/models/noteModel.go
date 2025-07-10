package models

import "gorm.io/gorm"

type Note struct {
	gorm.Model
	Text      string `json:"text"`
	TextColor string `json:"text_color"`
	BgColor   string `json:"bg_color"`
	TaskId    uint   `json:"task_id"`
	UserId    uint   `json:"user_id"`
}
