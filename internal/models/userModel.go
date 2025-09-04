package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Email      string  `json:"email" gorm:"unique"`
	Password   string  `json:"password"`
	ActiveTask *uint   `json:"active_task"`
	Theme      string  `json:"theme"`
	JobTitle   *string `json:"job_title"`
	TimeZone   *string `json:"time_zone"`
	Language   string  `json:"language"`
	Bio        string  `json:"bio"`
	Favorites  []uint  `json:"favorites" gorm:"serializer:json"`
	AvatarUrl  *string `json:"avatar_url"`
}
