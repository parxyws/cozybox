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

type PurchaseOrderRenderer struct {
	cfg *entity.Config
}

func NewPurchaseOrderRenderer(cfg *entity.Config) *PurchaseOrderRenderer {
	return &PurchaseOrderRenderer{cfg: cfg}
}

func (r *PurchaseOrderRenderer) Render(doc *models.Document) ([]byte, error) {
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

func (r *PurchaseOrderRenderer) buildProfileSection(doc *models.Document) []core.Row {
	color := &props.Color{Red: 186, Green: 216, Blue: 230} // Ice Blue

	outerTop := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}

	leftColMid := &props.Cell{BorderType: border.Left, BorderThickness: 0.2}
	rightColMid := &props.Cell{BorderType: border.Right, BorderThickness: 0.2}
	leftColBot := &props.Cell{BorderType: border.Left | border.Bottom, BorderThickness: 0.2}
	rightColBot := &props.Cell{BorderType: border.Right | border.Bottom, BorderThickness: 0.2}

	var rows []core.Row

	// Main Title
	rows = append(rows, row.New(0).WithStyle(outerTop).Add(
		text.NewCol(12, "PURCHASE ORDER", props.Text{Style: fontstyle.Bold, Size: 12, Align: align.Center, Bottom: 2, Top: 2}),
	))

	// Vendor and Ship To Headers
	rows = append(rows, row.New(5))
	rows = append(rows, row.New(0).Add(
		text.NewCol(5, "VENDOR", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}).WithStyle(outerTop),
		text.NewCol(1, "", props.Text{}),
		text.NewCol(6, "SHIP TO", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}).WithStyle(outerTop),
	))

	// Row 1: Name & PO Number
	rows = append(rows, row.New(0).Add(
		text.NewCol(1, "Name", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(leftColMid),
		text.NewCol(4, ": Main Supplier Corp", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(rightColMid),
		text.NewCol(1, "", props.Text{}),
		text.NewCol(2, "PO Number", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(leftColMid),
		text.NewCol(4, ": PO-2026-00100", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(rightColMid),
	))

	// Row 2: Address & Order Date
	rows = append(rows, row.New(0).Add(
		text.NewCol(1, "Address", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(leftColMid),
		text.NewCol(4, ": 123 Factory Line", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(rightColMid),
		text.NewCol(1, "", props.Text{}),
		text.NewCol(2, "Order Date", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(leftColMid),
		text.NewCol(4, ": 26 March 2026", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(rightColMid),
	))

	// Row 3: Phone & Company
	rows = append(rows, row.New(0).Add(
		text.NewCol(1, "Phone", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(leftColBot),
		text.NewCol(4, ": (555) 123-4567", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2, Bottom: 2}).WithStyle(rightColBot),
		text.NewCol(1, "", props.Text{}),
		text.NewCol(2, "Company", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2}).WithStyle(leftColBot),
		text.NewCol(4, ": PT. Cozybox Jaya Abadi", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2, Bottom: 2}).WithStyle(rightColBot),
	))

	return rows
}

func (r *PurchaseOrderRenderer) buildContentSection(items []models.DocumentItem) []core.Row {
	var rows []core.Row
	rows = append(rows, row.New(5))

	hdrCenter := props.Text{Style: fontstyle.Bold, Align: align.Center, Size: 8, Top: 2, Bottom: 2}
	hdrLeft := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}
	valCenter := props.Text{Align: align.Center, Size: 8, Top: 2, Bottom: 2}
	valLeft := props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}

	color := &props.Color{Red: 186, Green: 216, Blue: 230} // Ice Blue
	hdrCell := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}
	itemCell := &props.Cell{BorderType: border.Full, BorderThickness: 0.2}

	rows = append(rows, row.New(8).Add(
		text.NewCol(1, "#", hdrCenter).WithStyle(hdrCell),
		text.NewCol(6, "Item Specification", hdrLeft).WithStyle(hdrCell),
		text.NewCol(1, "Qty", hdrCenter).WithStyle(hdrCell),
		text.NewCol(2, "Estimated Cost", hdrCenter).WithStyle(hdrCell),
		text.NewCol(2, "Total Cost", hdrCenter).WithStyle(hdrCell),
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
		text.NewCol(2, "Total Authorized", summaryLabel).WithStyle(summaryCell),
		text.NewCol(2, "Rp 100.000.000", summaryValueBold).WithStyle(summaryCellRight),
	))

	// TERMS
	rows = append(rows, row.New(5))
	outerTop := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}
	outerBot := &props.Cell{BorderType: border.Left | border.Right | border.Bottom, BorderThickness: 0.2}

	rows = append(rows, row.New(8).WithStyle(outerTop).Add(
		text.NewCol(12, "PURCHASING CONDITIONS", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Left: 2}),
	))
	rows = append(rows, row.New(8).WithStyle(outerBot).Add(
		text.NewCol(12, "1. All goods must comply with provided specifications prior to shipping.", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2, Bottom: 2}),
	))

	return rows
}
