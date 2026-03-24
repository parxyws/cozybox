package models

import (
	"database/sql"
	"time"
)

type User struct {
	Id         string       `json:"id" gorm:"column:id"`
	Name       string       `json:"name" gorm:"column:name"`
	Username   string       `json:"username" gorm:"column:username"`
	Email      string       `json:"email" gorm:"column:email"`
	Password   string       `json:"password" gorm:"column:password"`
	IsVerified bool         `json:"is_verified" gorm:"column:is_verified"`
	LastLogin  time.Time    `json:"last_login" gorm:"column:last_login"`
	CreatedAt  time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  time.Time    `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt  sql.NullTime `json:"deleted_at" gorm:"column:deleted_at"`
}

func (u *User) TableName() string {
	return "users"
}
