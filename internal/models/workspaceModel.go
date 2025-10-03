package models

import "gorm.io/gorm"

type Workspace struct {
	gorm.Model
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ManagerId   uint   `json:"manager_id"`
	InviteCode  string `json:"invite_code"`
	Type        string `json:"type"`
}
