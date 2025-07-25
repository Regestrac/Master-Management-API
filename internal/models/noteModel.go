package models

import "gorm.io/gorm"

type Note struct {
	gorm.Model
	ID          uint   `json:"id"`
	Content     string `json:"content"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Width       uint   `json:"width"`
	Height      uint   `json:"height"`
	TextColor   string `json:"text_color"`
	BgColor     string `json:"bg_color"`
	BorderColor string `json:"border_color"`
	TaskId      uint   `json:"task_id"`
	UserId      uint   `json:"user_id"`
	Variant     string `json:"variant"`
}
