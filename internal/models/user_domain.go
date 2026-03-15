package models

import "database/sql"

type User struct {
	ID         string       `json:"id" gorm:"column:id"`
	Name       string       `json:"name" gorm:"column:name"`
	Username   string       `json:"username" gorm:"column:username"`
	Email      string       `json:"email" gorm:"column:email"`
	Password   string       `json:"password" gorm:"column:password"`
	IsVerified bool         `json:"is_verified" gorm:"column:is_verified"`
	CreatedAt  sql.NullTime `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  sql.NullTime `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt  sql.NullTime `json:"deleted_at" gorm:"column:deleted_at"`
}

func (u User) TableName() string {
	return "users"
}
