package dto

import "time"

type DocumentResponse struct {
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	DocumentType string    `json:"document_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
