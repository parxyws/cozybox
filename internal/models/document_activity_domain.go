package models

import "time"

// DocumentActivity records lifecycle events for audit trail.
type DocumentActivity struct {
	Id          string          `json:"id" gorm:"column:id;primaryKey"`
	DocumentId  string          `json:"document_id" gorm:"column:document_id"`
	Action      string          `json:"action" gorm:"column:action"`
	FromStatus  *DocumentStatus `json:"from_status" gorm:"column:from_status"`
	ToStatus    *DocumentStatus `json:"to_status" gorm:"column:to_status"`
	PerformedBy *string         `json:"performed_by" gorm:"column:performed_by"`
	Note        string          `json:"note" gorm:"column:note"`
	CreatedAt   time.Time       `json:"created_at" gorm:"column:created_at"`
}

func (da DocumentActivity) TableName() string {
	return "document_activities"
}
