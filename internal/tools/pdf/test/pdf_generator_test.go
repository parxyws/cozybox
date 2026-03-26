package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/johnfercher/maroto/v2"
	"github.com/parxyws/cozybox/internal/models"
	"github.com/parxyws/cozybox/internal/tools/pdf"
	"github.com/parxyws/cozybox/internal/tools/pdf/components"
	"github.com/shopspring/decimal"
)

func TestBuilderHeader(t *testing.T) {
	cfg := pdf.BuildConfig()
	m := maroto.New(cfg)
	m.RegisterHeader(components.BuildHeader()...)
	doc, err := m.Generate()
	if err != nil {
		t.Error(err)
	}
	doc.Save("header.pdf")
}

func TestBuilderFooter(t *testing.T) {
	cfg := pdf.BuildConfig()
	m := maroto.New(cfg)
	m.RegisterFooter(components.BuildFooter())
	doc, err := m.Generate()
	if err != nil {
		t.Error(err)
	}
	doc.Save("footer.pdf")
}

func TestBuilderDocument(t *testing.T) {
	cfg := pdf.BuildConfig()
	m := maroto.New(cfg)
	m.RegisterHeader(components.BuildHeader()...)
	m.RegisterFooter(components.BuildFooter())
	doc, err := m.Generate()
	if err != nil {
		t.Error(err)
	}
	doc.Save("document.pdf")
}

func TestBuilderQuotationDocument(t *testing.T) {
	cfg := pdf.BuildConfig()
	renderer := pdf.NewMarotoRenderer(cfg)

	var items []models.DocumentItem
	for i := 1; i <= 10; i++ {
		items = append(items, models.DocumentItem{
			Description: fmt.Sprintf("Service Item %d - Professional Integration", i),
			Quantity:    decimal.NewFromInt(int64((i % 3) + 1)),
			UnitPrice:   decimal.NewFromInt(int64(i * 1500000)),
		})
	}

	doc := &models.Document{
		Type:  models.DocumentTypeQuotation,
		Items: items,
	}

	docBytes, err := renderer.Generate(doc)
	if err != nil {
		t.Error(err)
	}

	err = os.WriteFile("quotation.pdf", docBytes, 0644)
	if err != nil {
		t.Error(err)
	}
}

func TestBuilderInvoiceDocument(t *testing.T) {
	cfg := pdf.BuildConfig()
	renderer := pdf.NewMarotoRenderer(cfg)

	var items []models.DocumentItem
	for i := 1; i <= 10; i++ {
		items = append(items, models.DocumentItem{
			Description: fmt.Sprintf("Service Item %d - Professional Integration", i),
			Quantity:    decimal.NewFromInt(int64((i % 3) + 1)),
			UnitPrice:   decimal.NewFromInt(int64(i * 1500000)),
		})
	}

	doc := &models.Document{
		Type:  models.DocumentTypeInvoice,
		Items: items,
	}

	docBytes, err := renderer.Generate(doc)
	if err != nil {
		t.Error(err)
	}

	err = os.WriteFile("invoice.pdf", docBytes, 0644)
	if err != nil {
		t.Error(err)
	}
}

func TestBuilderReceiptDocument(t *testing.T) {
	cfg := pdf.BuildConfig()
	renderer := pdf.NewMarotoRenderer(cfg)

	var items []models.DocumentItem
	for i := 1; i <= 10; i++ {
		items = append(items, models.DocumentItem{
			Description: fmt.Sprintf("Service Item %d - Professional Integration", i),
			Quantity:    decimal.NewFromInt(int64((i % 3) + 1)),
			UnitPrice:   decimal.NewFromInt(int64(i * 1500000)),
		})
	}

	doc := &models.Document{
		Type:  models.DocumentTypeReceipt,
		Items: items,
	}

	docBytes, err := renderer.Generate(doc)
	if err != nil {
		t.Error(err)
	}

	err = os.WriteFile("receipt.pdf", docBytes, 0644)
	if err != nil {
		t.Error(err)
	}
}

func TestBuilderPurchaseOrderDocument(t *testing.T) {
	cfg := pdf.BuildConfig()
	renderer := pdf.NewMarotoRenderer(cfg)

	var items []models.DocumentItem
	for i := 1; i <= 5; i++ {
		items = append(items, models.DocumentItem{
			Description: fmt.Sprintf("PO Item component %d", i),
			Quantity:    decimal.NewFromInt(int64(i * 2)),
			UnitPrice:   decimal.NewFromInt(int64(150000)),
		})
	}

	doc := &models.Document{
		Type:  models.DocumentTypePurchaseOrder,
		Items: items,
	}

	docBytes, err := renderer.Generate(doc)
	if err != nil {
		t.Error(err)
	}

	err = os.WriteFile("purchase_order.pdf", docBytes, 0644)
	if err != nil {
		t.Error(err)
	}
}

func TestBuilderSalesOrderDocument(t *testing.T) {
	cfg := pdf.BuildConfig()
	renderer := pdf.NewMarotoRenderer(cfg)

	var items []models.DocumentItem
	for i := 1; i <= 3; i++ {
		items = append(items, models.DocumentItem{
			Description: fmt.Sprintf("SO Ordered Goods %d", i),
			Quantity:    decimal.NewFromInt(int64(i)),
			UnitPrice:   decimal.NewFromInt(int64(5000000)),
		})
	}

	doc := &models.Document{
		Type:  models.DocumentTypeSalesOrder,
		Items: items,
	}

	docBytes, err := renderer.Generate(doc)
	if err != nil {
		t.Error(err)
	}

	err = os.WriteFile("sales_order.pdf", docBytes, 0644)
	if err != nil {
		t.Error(err)
	}
}

func TestBuilderDebitNoteDocument(t *testing.T) {
	cfg := pdf.BuildConfig()
	renderer := pdf.NewMarotoRenderer(cfg)

	var items []models.DocumentItem
	for i := 1; i <= 2; i++ {
		items = append(items, models.DocumentItem{
			Description: fmt.Sprintf("Price Diff Unit %d", i),
			Quantity:    decimal.NewFromInt(10),
			UnitPrice:   decimal.NewFromInt(50000),
		})
	}

	doc := &models.Document{
		Type:  models.DocumentTypeDebitNote,
		Items: items,
	}

	docBytes, err := renderer.Generate(doc)
	if err != nil {
		t.Error(err)
	}

	err = os.WriteFile("debit_note.pdf", docBytes, 0644)
	if err != nil {
		t.Error(err)
	}
}
