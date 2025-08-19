package models

import "gorm.io/gorm"

type Workspace struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	ManagerId   uint   `json:"manager_id"`
	InviteCode  string `json:"invite_code"`
	Type        string `json:"type"`
}
