package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// DocumentType represents the type discriminator for the unified documents table.
type DocumentType string

const (
	DocumentTypeQuotation     DocumentType = "quotation"
	DocumentTypeInvoice       DocumentType = "invoice"
	DocumentTypeReceipt       DocumentType = "receipt"
	DocumentTypePurchaseOrder DocumentType = "purchase_order"
	DocumentTypeSalesOrder    DocumentType = "sales_order"
	DocumentTypeDebitNote     DocumentType = "debit_note"
)

// DocumentStatus represents the lifecycle status of a document.
type DocumentStatus string

const (
	DocumentStatusDraft         DocumentStatus = "draft"
	DocumentStatusPublished     DocumentStatus = "published"
	DocumentStatusAccepted      DocumentStatus = "accepted"
	DocumentStatusRejected      DocumentStatus = "rejected"
	DocumentStatusPaid          DocumentStatus = "paid"
	DocumentStatusPartiallyPaid DocumentStatus = "partially_paid"
	DocumentStatusOverdue       DocumentStatus = "overdue"
	DocumentStatusCancelled     DocumentStatus = "cancelled"
)

// Document is the unified model for all 6 document types:
// quotation, invoice, receipt, purchase_order, sales_order, debit_note.
type Document struct {
	Id             string  `json:"id" gorm:"column:id;primaryKey"`
	TenantId       string  `json:"tenant_id" gorm:"column:tenant_id;index"`
	OrganizationId string  `json:"organization_id" gorm:"column:organization_id"`
	ContactId      *string `json:"contact_id" gorm:"column:contact_id"`
	ParentId       *string `json:"parent_id" gorm:"column:parent_id"`

	// Classification
	Type        DocumentType   `json:"type" gorm:"column:type"`
	Status      DocumentStatus `json:"status" gorm:"column:status;default:draft"`
	DocumentRef string         `json:"document_ref" gorm:"column:document_ref"`

	// Dates
	IssueDate  sql.NullTime `json:"issue_date" gorm:"column:issue_date"`
	DueDate    sql.NullTime `json:"due_date" gorm:"column:due_date"`
	ValidUntil sql.NullTime `json:"valid_until" gorm:"column:valid_until"`

	// Financial
	Currency       string          `json:"currency" gorm:"column:currency;default:USD"`
	Subtotal       decimal.Decimal `json:"subtotal" gorm:"column:subtotal;default:0"`
	DiscountAmount decimal.Decimal `json:"discount_amount" gorm:"column:discount_amount;default:0"`
	TaxAmount      decimal.Decimal `json:"tax_amount" gorm:"column:tax_amount;default:0"`
	Total          decimal.Decimal `json:"total" gorm:"column:total;default:0"`
	AmountPaid     decimal.Decimal `json:"amount_paid" gorm:"column:amount_paid;default:0"`

	// Content
	Notes  string `json:"notes" gorm:"column:notes"`
	Terms  string `json:"terms" gorm:"column:terms"`
	Footer string `json:"footer" gorm:"column:footer"`

	// Type-specific extensible metadata stored as JSONB
	Metadata json.RawMessage `json:"metadata" gorm:"column:metadata;type:jsonb;default:'{}'"`

	// File storage
	PdfUrl string `json:"pdf_url" gorm:"column:pdf_url"`

	// Audit
	CreatedBy string       `json:"created_by" gorm:"column:created_by"`
	CreatedAt time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time    `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" gorm:"column:deleted_at"`

	// Relations
	Tenant       Tenant             `json:"tenant" gorm:"foreignKey:TenantId;references:Id"`
	Organization Organization       `json:"organization" gorm:"foreignKey:OrganizationId;references:Id"`
	Contact      *Contact           `json:"contact" gorm:"foreignKey:ContactId;references:Id"`
	Parent       *Document          `json:"parent" gorm:"foreignKey:ParentId;references:Id"`
	Children     []Document         `json:"children" gorm:"foreignKey:ParentId;references:Id"`
	Items        []DocumentItem     `json:"items" gorm:"foreignKey:DocumentId;references:Id"`
	Activities   []DocumentActivity `json:"activities" gorm:"foreignKey:DocumentId;references:Id"`
	Creator      User               `json:"creator" gorm:"foreignKey:CreatedBy;references:Id"`
}

func (d Document) TableName() string {
	return "documents"
}

// IsFinanciallySettled returns true if the document is fully paid.
func (d Document) IsFinanciallySettled() bool {
	return d.AmountPaid.GreaterThanOrEqual(d.Total)
}

// OutstandingAmount returns the remaining balance to be paid.
func (d Document) OutstandingAmount() decimal.Decimal {
	return d.Total.Sub(d.AmountPaid)
}
