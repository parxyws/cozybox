package template

import (
	"fmt"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/parxyws/cozybox/internal/models"
	"github.com/parxyws/cozybox/internal/tools/pdf/components"
)

type SalesOrderRenderer struct {
	cfg *entity.Config
}

func NewSalesOrderRenderer(cfg *entity.Config) *SalesOrderRenderer {
	return &SalesOrderRenderer{cfg: cfg}
}

func (r *SalesOrderRenderer) Render(doc *models.Document) ([]byte, error) {
	m := maroto.New(r.cfg)
	m.RegisterHeader(components.BuildHeader()...)
	m.RegisterFooter(components.BuildFooter())

	m.AddRows(r.buildProfileSection(doc)...)
	m.AddRows(r.buildContentSection(doc.Items)...)

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}
	return document.GetBytes(), nil
}

func (r *SalesOrderRenderer) buildProfileSection(doc *models.Document) []core.Row {
	color := &props.Color{Red: 255, Green: 204, Blue: 153} // Peach/Orange

	outerTop := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}
	outerMid := &props.Cell{BorderType: border.Left | border.Right, BorderThickness: 0.2}
	outerBot := &props.Cell{BorderType: border.Left | border.Right | border.Bottom, BorderThickness: 0.2}

	var rows []core.Row

	// Main Title
	rows = append(rows, row.New(0).WithStyle(outerTop).Add(
		text.NewCol(12, "SALES ORDER", props.Text{Style: fontstyle.Bold, Size: 12, Align: align.Center, Bottom: 2, Top: 2}),
	))

	// Trifold Layout Headers
	rows = append(rows, row.New(5))
	rows = append(rows, row.New(0).Add(
		text.NewCol(4, "CUSTOMER", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}).WithStyle(outerTop),
		text.NewCol(4, "ORDER DETAILS", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}).WithStyle(outerTop),
		text.NewCol(4, "SHIPPING INFO", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}).WithStyle(outerTop),
	))

	// Trifold Layout Mid Row
	rows = append(rows, row.New(0).Add(
		text.NewCol(4, "Perkakas Rekadaya Nusantara", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(outerMid),
		text.NewCol(4, "Order No: SO-2026-991", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(outerMid),
		text.NewCol(4, "Carrier: Express Logistic", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(outerMid),
	))

	// Trifold Layout Bot Row
	rows = append(rows, row.New(0).Add(
		text.NewCol(4, "62-21-6515551 Ext. 61909", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2, Bottom: 2}).WithStyle(outerBot),
		text.NewCol(4, "Order Date: 26 March 2026", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2, Bottom: 2}).WithStyle(outerBot),
		text.NewCol(4, "Expected: 30 March 2026", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2, Bottom: 2}).WithStyle(outerBot),
	))

	return rows
}

func (r *SalesOrderRenderer) buildContentSection(items []models.DocumentItem) []core.Row {
	var rows []core.Row
	rows = append(rows, row.New(5))

	hdrCenter := props.Text{Style: fontstyle.Bold, Align: align.Center, Size: 8, Top: 2, Bottom: 2}
	hdrLeft := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}
	valCenter := props.Text{Align: align.Center, Size: 8, Top: 2, Bottom: 2}
	valLeft := props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}

	color := &props.Color{Red: 255, Green: 204, Blue: 153} // Peach/Orange
	hdrCell := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}
	itemCell := &props.Cell{BorderType: border.Full, BorderThickness: 0.2}

	rows = append(rows, row.New(8).Add(
		text.NewCol(1, "#", hdrCenter).WithStyle(hdrCell),
		text.NewCol(6, "Description", hdrLeft).WithStyle(hdrCell),
		text.NewCol(1, "Qty", hdrCenter).WithStyle(hdrCell),
		text.NewCol(2, "Unit Price", hdrCenter).WithStyle(hdrCell),
		text.NewCol(2, "Amount", hdrCenter).WithStyle(hdrCell),
	))

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

	// TERMS
	rows = append(rows, row.New(5))
	outerTop := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}
	outerBot := &props.Cell{BorderType: border.Left | border.Right | border.Bottom, BorderThickness: 0.2}

	rows = append(rows, row.New(8).WithStyle(outerTop).Add(
		text.NewCol(12, "SALES TERMS", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Left: 2}),
	))
	rows = append(rows, row.New(8).WithStyle(outerBot).Add(
		text.NewCol(12, "1. This order constitutes binding agreement to the delivery terms.", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2, Bottom: 2}),
	))

	return rows
}
