package service

import (
	"github.com/parxyws/cozybox/internal/dto"
	"gorm.io/gorm"
)

type DocumentService struct {
	db *gorm.DB
}

func NewDocumentService(db *gorm.DB) *DocumentService {
	return &DocumentService{db: db}
}

func (d *DocumentService) CreateDocument(req *dto.CreateDocumentRequest) (*dto.DocumentResponse, error) {
	
	return &dto.DocumentResponse{}, nil
}
