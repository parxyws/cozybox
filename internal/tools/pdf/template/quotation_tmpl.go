package template

import (
	"fmt"

	"github.com/johnfercher/maroto/v2"
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

type QuotationRenderer struct {
	cfg *entity.Config
}

func NewQuotationRenderer(cfg *entity.Config) *QuotationRenderer {
	return &QuotationRenderer{cfg: cfg}
}

func (q *QuotationRenderer) Render(doc *models.Document) ([]byte, error) {
	m := maroto.New(q.cfg)

	m.RegisterHeader(components.BuildHeader()...)
	m.RegisterFooter(components.BuildFooter())

	m.AddRows(q.buildQuotationProfileSection(doc)...)
	m.AddRows(q.buildQuotationContentSection(doc.Items)...)

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return document.GetBytes(), nil
}

func (q *QuotationRenderer) buildQuotationProfileSection(doc *models.Document) []core.Row {
	padding := 2.0

	lblTop := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: padding, Left: padding}
	valTop := props.Text{Align: align.Left, Size: 8, Top: padding}

	lblMid := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Left: padding}
	valMid := props.Text{Align: align.Left, Size: 8}

	lblBot := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Bottom: padding, Left: padding}
	valBot := props.Text{Align: align.Left, Size: 8, Bottom: padding}

	sep := props.Line{Orientation: orientation.Vertical, Thickness: 0.2, SizePercent: 100, OffsetPercent: 100}

	color := &props.Color{Red: 173, Green: 181, Blue: 189}

	return []core.Row{
		// Title
		row.New(0).WithStyle(&props.Cell{BorderType: border.Top | border.Left | border.Right, BorderThickness: 0.2, BackgroundColor: color}).Add(
			text.NewCol(12, "QUOTATION: PT. TOYOTA MOTOR MANUFACTURING INDONESIA", props.Text{Style: fontstyle.Bold, Size: 12, Align: align.Center, Bottom: padding + 1, Top: padding}),
		),

		// Row 1
		row.New(0).WithStyle(&props.Cell{BorderType: border.Top | border.Left | border.Right, BorderThickness: 0.2}).Add(
			text.NewCol(2, "To", lblTop),
			text.NewCol(4, ": Perkakas Rekadaya Nusantara", valTop),
			line.NewCol(1, sep),
			text.NewCol(2, "Phone", lblTop),
			text.NewCol(3, ": 62-21-6515551 Ext. 61909", valTop),
		),
		// Row 2
		row.New(0).WithStyle(&props.Cell{BorderType: border.Left | border.Right, BorderThickness: 0.2}).Add(
			text.NewCol(2, "Attn.", lblMid),
			text.NewCol(4, ": Mr. Agustian Sabar L", valMid),
			line.NewCol(1, sep),
			text.NewCol(2, "Ref. No.", lblMid),
			text.NewCol(3, ": PBMD/EXT-COM/2218/VII/2025", valMid),
		),
		// Row 3
		row.New(0).WithStyle(&props.Cell{BorderType: border.Left | border.Right | border.Bottom, BorderThickness: 0.2}).Add(
			text.NewCol(2, "Cc", lblBot),
			text.NewCol(4, ": Mr. Afdely Sidqy", valBot),
			line.NewCol(1, sep),
			text.NewCol(2, "Issued Date", lblBot),
			text.NewCol(3, ": 28 July 2025", valBot),
		),
	}
}

func (q *QuotationRenderer) buildQuotationContentSection(items []models.DocumentItem) []core.Row {
	var rows []core.Row

	// Initial top spacing
	rows = append(rows, row.New(5))

	// Styles
	hdrCenter := props.Text{Style: fontstyle.Bold, Align: align.Center, Size: 8, Top: 2, Bottom: 2}
	hdrLeft := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}

	valCenter := props.Text{Align: align.Center, Size: 8, Top: 2, Bottom: 2}
	valLeft := props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}
	// valRight := props.Text{Align: align.Right, Size: 8, Top: 2, Bottom: 2, Right: 2}

	color := &props.Color{Red: 173, Green: 181, Blue: 189}

	// Cell border styles for the table
	hdrCell := &props.Cell{
		BorderType:      border.Full,
		BorderThickness: 0.2,
		BackgroundColor: color,
	}

	itemCell := &props.Cell{
		BorderType:      border.Full,
		BorderThickness: 0.2,
	}

	// Table Header (Each cell bordered individually + Gray Background)
	rows = append(rows, row.New(0).Add(
		text.NewCol(1, "#", hdrCenter).WithStyle(hdrCell),
		text.NewCol(6, "Description", hdrLeft).WithStyle(hdrCell),
		text.NewCol(1, "Qty", hdrCenter).WithStyle(hdrCell),
		text.NewCol(2, "Unit Price", hdrCenter).WithStyle(hdrCell),
		text.NewCol(2, "Total Price", hdrCenter).WithStyle(hdrCell),
	))

	// Table Content (Each cell bordered individually)
	for i, item := range items {
		rows = append(rows, row.New(0).Add(
			text.NewCol(1, fmt.Sprintf("%d", i+1), valCenter).WithStyle(itemCell),
			text.NewCol(6, item.Description, valLeft).WithStyle(itemCell),
			text.NewCol(1, item.Quantity.String(), valCenter).WithStyle(itemCell),
			text.NewCol(2, fmt.Sprintf("Rp %s", item.UnitPrice.StringFixed(2)), valLeft).WithStyle(itemCell),
			text.NewCol(2, fmt.Sprintf("Rp %s", item.Quantity.Mul(item.UnitPrice).Round(2).StringFixed(2)), valLeft).WithStyle(itemCell),
		))
	}

	rows = append(rows, row.New(5))

	// Cell border style for summary block
	summaryCell := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}
	summaryCellRight := &props.Cell{BorderType: border.Full, BorderThickness: 0.2}

	// Subtotal, Tax, Total
	rows = append(rows, row.New(0).Add(
		text.NewCol(8, "", props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}),
		text.NewCol(2, "Subtotal", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}).WithStyle(summaryCell),
		text.NewCol(2, "Rp 100.000.000", props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}).WithStyle(summaryCellRight),
	))

	rows = append(rows, row.New(0).Add(
		text.NewCol(8, "", props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}),
		text.NewCol(2, "Tax (11%)", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}).WithStyle(summaryCell),
		text.NewCol(2, "Rp 11.000.000", props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}).WithStyle(summaryCellRight),
	))

	rows = append(rows, row.New(0).Add(
		text.NewCol(8, "", props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}),
		text.NewCol(2, "Total", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}).WithStyle(summaryCell),
		text.NewCol(2, "Rp 111.000.000", props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}).WithStyle(summaryCellRight),
	))

	// Styles for NOTE and TERMS bounding boxes
	outerTop := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}
	outerMid := &props.Cell{BorderType: border.Left | border.Right, BorderThickness: 0.2}
	outerBot := &props.Cell{BorderType: border.Left | border.Right | border.Bottom, BorderThickness: 0.2}

	// NOTE Section
	rows = append(rows, row.New(5))

	rows = append(rows, row.New(0).WithStyle(outerTop).Add(
		text.NewCol(12, "NOTE", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}),
	))
	rows = append(rows, row.New(0).WithStyle(outerMid).Add(
		text.NewCol(12, "A. ", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}),
	))
	rows = append(rows, row.New(0).WithStyle(outerMid).Add(
		text.NewCol(12, "B. ", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}),
	))
	rows = append(rows, row.New(0).WithStyle(outerMid).Add(
		text.NewCol(12, "C. ", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}),
	))
	rows = append(rows, row.New(0).WithStyle(outerMid).Add(
		text.NewCol(12, "D. ", props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}),
	))
	rows = append(rows, row.New(0).WithStyle(outerBot).Add(
		text.NewCol(12, " ", props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}),
	))

	// TERMS AND CONDITIONS Section
	rows = append(rows, row.New(5))
	rows = append(rows, row.New(0).WithStyle(outerTop).Add(
		text.NewCol(12, "TERMS AND CONDITIONS", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}),
	))
	rows = append(rows, row.New(0).WithStyle(outerMid).Add(
		text.NewCol(12, "1. This quotation is valid for 30 days from the date of issue.", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}),
	))
	rows = append(rows, row.New(0).WithStyle(outerMid).Add(
		text.NewCol(12, "2. Payment is required in full within 14 days of invoice receipt.", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}),
	))
	rows = append(rows, row.New(0).WithStyle(outerMid).Add(
		text.NewCol(12, "3. All prices are strictly exclusive of VAT and additional taxes.", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}),
	))
	rows = append(rows, row.New(0).WithStyle(outerBot).Add(
		text.NewCol(12, "4. Execution of services will commence upon receipt of a signed approval or PO.", props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}),
	))

	return rows
}
