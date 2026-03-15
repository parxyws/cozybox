package models

import "database/sql"

type Quotation struct {
	Id          int64        `json:"id" gorm:"column:id"`
	Name        string       `json:"name" gorm:"column:name"`
	DocumentRef string       `json:"document_ref" gorm:"column:document_ref"`
	Date        sql.NullTime `json:"date" gorm:"column:date"`
	DueDate     sql.NullTime `json:"due_date" gorm:"column:due_date"`
	Currency    string       `json:"currency" gorm:"column:currency"`
	Subtotal    int          `json:"subtotal" gorm:"column:subtotal"`
	Tax         int          `json:"tax" gorm:"column:tax"`
	Total       int          `json:"total" gorm:"column:total"`
	Notes       string       `json:"notes" gorm:"column:notes"`
	CreatedAt   sql.NullTime `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   sql.NullTime `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt   sql.NullTime `json:"deleted_at" gorm:"column:deleted_at"`

	Records []Record `json:"records" gorm:"foreignKey:DocumentId;references:Id"`
}

func (q Quotation) TableName() string {
	return "quotations"
}
