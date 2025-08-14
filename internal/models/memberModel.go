package models

import (
	"time"

	"gorm.io/gorm"
)

type Member struct {
	gorm.Model
	WorkspaceId  uint       `json:"workspace_id"`
	UserId       uint       `json:"user_id"`
	Role         string     `json:"role"`
	JoinedAt     *time.Time `json:"joined_at"`
	AvatarUrl    string     `json:"avatar_url"`
	ProfileColor string     `json:"profile_color"`
}
