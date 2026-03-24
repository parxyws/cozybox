package models

// DocumentSequence manages auto-incrementing document reference numbers
// per organization and document type.
type DocumentSequence struct {
	Id             string       `json:"id" gorm:"column:id;primaryKey"`
	OrganizationId string       `json:"organization_id" gorm:"column:organization_id"`
	Type           DocumentType `json:"type" gorm:"column:type"`
	Prefix         string       `json:"prefix" gorm:"column:prefix"`
	NextNumber     int          `json:"next_number" gorm:"column:next_number;default:1"`
	Format         string       `json:"format" gorm:"column:format;default:{PREFIX}-{YEAR}-{SEQ:4}"`
}

func (ds DocumentSequence) TableName() string {
	return "document_sequences"
}
