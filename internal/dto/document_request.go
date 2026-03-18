package dto

type CreateDocumentRequest struct {
	Name         string `json:"name"         validate:"required,min=2,max=100"`
	DocumentType string `json:"document_type" validate:"required"`
}

type UpdateDocumentRequest struct {
	Name         string `json:"name"         validate:"omitempty,min=2,max=100"`
	DocumentType string `json:"document_type" validate:"omitempty"`
}

type GetDocumentRequest struct {
	Id string `json:"id" validate:"required"`
}

type DeleteDocumentRequest struct {
	Id string `json:"id" validate:"required"`
}
