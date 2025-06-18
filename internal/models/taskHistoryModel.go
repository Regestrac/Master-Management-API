package models

import "gorm.io/gorm"

type TaskHistory struct {
	gorm.Model
	Action string `json:"action"` // "status_update" | "title_update" | "desc_update" | "started" | "stopped" | "created" | "note" | "subtask" | "checklist"
	Before string `json:"before"`
	After  string `json:"after"`
	TaskId uint   `json:"task_id"`
	UserId uint   `json:"user_id"` // ID of the user who performed the action
}
