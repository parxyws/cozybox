package models

import (
	"database/sql"

	"github.com/shopspring/decimal"
)

type Record struct {
	Id         string          `json:"id" gorm:"column:id"`
	Item       string          `json:"item" gorm:"column:item"`
	Quantity   int             `json:"quantity" gorm:"column:quantity"`
	Price      decimal.Decimal `json:"price" gorm:"column:price"`
	Amount     int             `json:"amount" gorm:"column:amount"`
	Currency   string          `json:"currency" gorm:"column:currency"`
	DocumentId string          `json:"document_id" gorm:"column:document_id"`
	CreatedAt  sql.NullTime    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  sql.NullTime    `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt  sql.NullTime    `json:"deleted_at" gorm:"column:deleted_at"`
}

func (r Record) TableName() string {
	return "records"
}
