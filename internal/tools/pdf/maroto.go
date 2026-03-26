package pdf

import (
	"fmt"

	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/parxyws/cozybox/internal/models"
	"github.com/parxyws/cozybox/internal/tools/pdf/template"
)

type DocumentRenderer interface {
	Render(doc *models.Document) ([]byte, error)
}

type MarotoRenderer struct {
	config   *entity.Config
	renderer map[models.DocumentType]DocumentRenderer
}

func NewMarotoRenderer(config *entity.Config) *MarotoRenderer {
	cfg := BuildConfig()
	g := &MarotoRenderer{
		config:   cfg,
		renderer: make(map[models.DocumentType]DocumentRenderer),
	}

	g.renderer[models.DocumentTypeQuotation] = template.NewQuotationRenderer(cfg)
	g.renderer[models.DocumentTypeInvoice] = template.NewInvoiceRenderer(cfg)
	g.renderer[models.DocumentTypeReceipt] = template.NewReceiptRenderer(cfg)
	g.renderer[models.DocumentTypePurchaseOrder] = template.NewPurchaseOrderRenderer(cfg)
	g.renderer[models.DocumentTypeSalesOrder] = template.NewSalesOrderRenderer(cfg)
	g.renderer[models.DocumentTypeDebitNote] = template.NewDebitNoteRenderer(cfg)

	return g
}

func (g *MarotoRenderer) Generate(doc *models.Document) ([]byte, error) {
	renderer, ok := g.renderer[doc.Type]
	if !ok {
		return nil, fmt.Errorf("renderer not found for document type %s", doc.Type)
	}
	return renderer.Render(doc)
}
