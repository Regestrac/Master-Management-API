package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email" gorm:"unique"`
	Password   string `json:"password"`
	ActiveTask *uint  `json:"active_task"`
	Theme      string `json:"theme"`
}
