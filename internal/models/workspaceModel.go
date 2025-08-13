package models

type Workspace struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerId     uint   `json:"owner_id"`
	InviteCode  string `json:"invite_code"`
}
