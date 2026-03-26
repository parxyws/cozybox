package template

import (
	"fmt"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/parxyws/cozybox/internal/models"
	"github.com/parxyws/cozybox/internal/tools/pdf/components"
)

type ReceiptRenderer struct {
	cfg *entity.Config
}

func NewReceiptRenderer(cfg *entity.Config) *ReceiptRenderer {
	return &ReceiptRenderer{cfg: cfg}
}

func (r *ReceiptRenderer) Render(doc *models.Document) ([]byte, error) {
	m := maroto.New(r.cfg)

	m.RegisterHeader(components.BuildHeader()...)
	m.RegisterFooter(components.BuildFooter())

	m.AddRows(r.buildReceiptProfileSection(doc)...)
	m.AddRows(r.buildReceiptContentSection(doc.Items)...)

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return document.GetBytes(), nil
}

func (r *ReceiptRenderer) buildReceiptProfileSection(doc *models.Document) []core.Row {
	padding := 1.0

	lblTop := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: padding, Left: padding}
	valTop := props.Text{Align: align.Left, Size: 8, Top: padding}

	lblMid := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Left: padding}
	valMid := props.Text{Align: align.Left, Size: 8}

	lblBot := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Bottom: padding, Left: padding}
	valBot := props.Text{Align: align.Left, Size: 8, Bottom: padding}

	sep := props.Line{Orientation: orientation.Vertical, Thickness: 0.2, SizePercent: 100, OffsetPercent: 100}

	color := &props.Color{Red: 181, Green: 201, Blue: 154}

	return []core.Row{
		// Title
		row.New(0).WithStyle(&props.Cell{BorderType: border.Top | border.Left | border.Right, BorderThickness: 0.2, BackgroundColor: color}).Add(
			text.NewCol(12, "RECEIPT", props.Text{Style: fontstyle.Bold, Size: 12, Align: align.Center, Bottom: padding + 1, Top: padding}),
		),

		// Row 1
		row.New(0).WithStyle(&props.Cell{BorderType: border.Top | border.Left | border.Right, BorderThickness: 0.2}).Add(
			text.NewCol(2, "Bill To", lblTop),
			text.NewCol(4, ": Perkakas Rekadaya Nusantara", valTop),
			line.NewCol(1, sep),
			text.NewCol(2, "Receipt No", lblTop),
			text.NewCol(3, ": PBMD/EXT-COM/2218/VII/2025", valTop),
		),
		// Row 2
		row.New(0).WithStyle(&props.Cell{BorderType: border.Left | border.Right, BorderThickness: 0.2}).Add(
			text.NewCol(2, "Attn.", lblMid),
			text.NewCol(4, ": Mr. Agustian Sabar L", valMid),
			line.NewCol(1, sep),
			text.NewCol(2, "Issue Date", lblMid),
			text.NewCol(3, ": 28 July 2025", valMid),
		),
		// Row 3
		row.New(0).WithStyle(&props.Cell{BorderType: border.Left | border.Right, BorderThickness: 0.2}).Add(
			text.NewCol(2, "Phone", lblMid),
			text.NewCol(4, ": 62-21-6515551 Ext. 61909", valMid),
			line.NewCol(1, sep),
			text.NewCol(2, "Payment Date", lblMid),
			text.NewCol(3, ": 28 July 2025", valMid),
		),

		row.New(0).WithStyle(&props.Cell{BorderType: border.Left | border.Right | border.Bottom, BorderThickness: 0.2}).Add(
			text.NewCol(2, "", lblBot),
			text.NewCol(4, "", valBot),
			line.NewCol(1, sep),
			text.NewCol(2, "Invoice No", lblBot),
			text.NewCol(3, ": PO/123456789", valBot),
		),
	}
}

func (r *ReceiptRenderer) buildReceiptContentSection(items []models.DocumentItem) []core.Row {
	var rows []core.Row

	rows = append(rows, row.New(5))

	// Styles
	hdrCenter := props.Text{Style: fontstyle.Bold, Align: align.Center, Size: 8, Top: 2, Bottom: 2}
	hdrLeft := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}
	valCenter := props.Text{Align: align.Center, Size: 8, Top: 2, Bottom: 2}
	valLeft := props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}

	color := &props.Color{Red: 181, Green: 201, Blue: 154}

	hdrCell := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}
	itemCell := &props.Cell{BorderType: border.Full, BorderThickness: 0.2}

	// ── Table Header ─────────────────────────────────────────────────
	rows = append(rows, row.New(8).Add(
		text.NewCol(1, "#", hdrCenter).WithStyle(hdrCell),
		text.NewCol(6, "Description", hdrLeft).WithStyle(hdrCell),
		text.NewCol(1, "Qty", hdrCenter).WithStyle(hdrCell),
		text.NewCol(2, "Unit Price", hdrCenter).WithStyle(hdrCell),
		text.NewCol(2, "Total Price", hdrCenter).WithStyle(hdrCell),
	))

	// ── Table Rows ───────────────────────────────────────────────────
	for i, item := range items {
		rows = append(rows, row.New(8).Add(
			text.NewCol(1, fmt.Sprintf("%d", i+1), valCenter).WithStyle(itemCell),
			text.NewCol(6, item.Description, valLeft).WithStyle(itemCell),
			text.NewCol(1, item.Quantity.String(), valCenter).WithStyle(itemCell),
			text.NewCol(2, fmt.Sprintf("Rp %s", item.UnitPrice.StringFixed(2)), valLeft).WithStyle(itemCell),
			text.NewCol(2, fmt.Sprintf("Rp %s", item.Quantity.Mul(item.UnitPrice).Round(2).StringFixed(2)), valLeft).WithStyle(itemCell),
		))
	}

	rows = append(rows, row.New(5))

	// ── Summary block ────────────────────────────────────────────────
	summaryCell := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}
	summaryCellRight := &props.Cell{BorderType: border.Full, BorderThickness: 0.2}

	summaryLabel := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}
	summaryValue := props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}
	summaryValueBold := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}
	emptyCol := props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}

	rows = append(rows, row.New(8).Add(
		text.NewCol(8, "", emptyCol),
		text.NewCol(2, "Subtotal", summaryLabel).WithStyle(summaryCell),
		text.NewCol(2, "Rp 100.000.000", summaryValue).WithStyle(summaryCellRight),
	))
	rows = append(rows, row.New(8).Add(
		text.NewCol(8, "", emptyCol),
		text.NewCol(2, "Tax (11%)", summaryLabel).WithStyle(summaryCell),
		text.NewCol(2, "Rp 11.000.000", summaryValue).WithStyle(summaryCellRight),
	))
	rows = append(rows, row.New(8).Add(
		text.NewCol(8, "", emptyCol),
		text.NewCol(2, "Total", summaryLabel).WithStyle(summaryCell),
		text.NewCol(2, "Rp 111.000.000", summaryValueBold).WithStyle(summaryCellRight),
	))
	rows = append(rows, row.New(8).Add(
		text.NewCol(8, "", emptyCol),
		text.NewCol(2, "AMOUNT PAID", summaryLabel).WithStyle(summaryCell),
		text.NewCol(2, "Rp 111.000.000", summaryValueBold).WithStyle(summaryCellRight),
	))

	// ── Payment section border styles ────────────────────────────────
	outerTop := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}
	outerMidTop := &props.Cell{BorderType: border.Left | border.Right | border.Top, BorderThickness: 0.2}
	outerMid := &props.Cell{BorderType: border.Left | border.Right, BorderThickness: 0.2}
	outerBot := &props.Cell{BorderType: border.Left | border.Right | border.Bottom, BorderThickness: 0.2}
	leftColMid := &props.Cell{BorderType: border.Left, BorderThickness: 0.2}
	rightColMid := &props.Cell{BorderType: border.Right, BorderThickness: 0.2}
	leftColBot := &props.Cell{BorderType: border.Left | border.Bottom, BorderThickness: 0.2}
	rightColBot := &props.Cell{BorderType: border.Right | border.Bottom, BorderThickness: 0.2}

	// ── Payment section sizing ───────────────────────────────────────
	const (
		paymentRowHeight = float64(5) // reduced from 6 to match tighter padding
		paymentNumRows   = float64(4)
	)
	totalContentHeight := paymentRowHeight * paymentNumRows // 20
	statusTextTop := (totalContentHeight / 2) - 4           // vertically center size-11 text

	rows = append(rows, row.New(5))

	// ── Payment Headers ──────────────────────────────────────────────
	rows = append(rows, row.New(6).Add(
		text.NewCol(7, "PAYMENT DETAILS", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 1, Bottom: 1, Left: 2}).WithStyle(outerTop),
		col.New(1),
		text.NewCol(4, "PAYMENT STATUS", props.Text{Style: fontstyle.Bold, Align: align.Center, Size: 8, Top: 1, Bottom: 1}).WithStyle(outerTop),
	))

	// ── Row 1: Bank | Status text ──
	rows = append(rows, row.New(paymentRowHeight).Add(
		text.NewCol(2, "Bank", props.Text{Align: align.Left, Size: 8, Top: 1, Bottom: 1, Left: 2}).WithStyle(leftColMid),
		text.NewCol(5, ": Bank Central Asia (BCA)", props.Text{Align: align.Left, Size: 8, Top: 1, Bottom: 1, Left: 2}).WithStyle(rightColMid),
		col.New(1),
		text.NewCol(4, "RECEIVED / PAID", props.Text{
			Style: fontstyle.Bold,
			Align: align.Center,
			Size:  12,
			Top:   statusTextTop,
		}).WithStyle(outerMidTop),
	))

	// ── Row 2: Account Name ──────────────────────────────────────────
	rows = append(rows, row.New(paymentRowHeight).Add(
		text.NewCol(2, "Account Name", props.Text{Align: align.Left, Size: 8, Top: 1, Bottom: 1, Left: 2}).WithStyle(leftColMid),
		text.NewCol(5, ": PT. Cozybox Jaya Abadi", props.Text{Align: align.Left, Size: 8, Top: 1, Bottom: 1, Left: 2}).WithStyle(rightColMid),
		col.New(1),
		col.New(4).WithStyle(outerMid),
	))

	// ── Row 3: Account Number ────────────────────────────────────────
	rows = append(rows, row.New(paymentRowHeight).Add(
		text.NewCol(2, "Account Number", props.Text{Align: align.Left, Size: 8, Top: 1, Bottom: 1, Left: 2}).WithStyle(leftColMid),
		text.NewCol(5, ": 1234567890", props.Text{Align: align.Left, Size: 8, Top: 1, Bottom: 1, Left: 2}).WithStyle(rightColMid),
		col.New(1),
		col.New(4).WithStyle(outerMid),
	))

	// ── Row 4: Reference | box closes ────────────────────────────────
	rows = append(rows, row.New(paymentRowHeight).Add(
		text.NewCol(2, "Reference", props.Text{Align: align.Left, Size: 8, Top: 1, Bottom: 1, Left: 2}).WithStyle(leftColBot),
		text.NewCol(5, ": PBMD/EXT-COM/2218/VII/2025", props.Text{Align: align.Left, Size: 8, Top: 1, Bottom: 1, Left: 2}).WithStyle(rightColBot),
		col.New(1),
		col.New(4).WithStyle(outerBot),
	))

	// ── Terms ────────────────────────────────────────────────────────
	rows = append(rows, row.New(5))
	rows = append(rows, row.New(8).WithStyle(outerTop).Add(
		text.NewCol(12, "TERMS", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Left: 2}),
	))
	rows = append(rows, row.New(8).WithStyle(outerBot).Add(
		text.NewCol(12, "1. This document serves as official proof of payment for the transaction.", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2, Bottom: 2}),
	))

	return rows
}
