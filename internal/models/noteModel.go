package models

import "gorm.io/gorm"

type Note struct {
	gorm.Model
	ID        uint   `json:"id"`
	Text      string `json:"text"`
	TextColor string `json:"text_color"`
	BgColor   string `json:"bg_color"`
	TaskId    uint   `json:"task_id"`
	UserId    uint   `json:"user_id"`
}
