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

type DebitNoteRenderer struct {
	cfg *entity.Config
}

func NewDebitNoteRenderer(cfg *entity.Config) *DebitNoteRenderer {
	return &DebitNoteRenderer{cfg: cfg}
}

func (r *DebitNoteRenderer) Render(doc *models.Document) ([]byte, error) {
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

func (r *DebitNoteRenderer) buildProfileSection(doc *models.Document) []core.Row {
	padding := 1.0

	lblTop := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: padding, Left: padding}
	valTop := props.Text{Align: align.Left, Size: 8, Top: padding}
	lblMid := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Left: padding}
	valMid := props.Text{Align: align.Left, Size: 8}
	lblBot := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Bottom: padding, Left: padding}
	valBot := props.Text{Align: align.Left, Size: 8, Bottom: padding}

	sep := props.Line{Orientation: orientation.Vertical, Thickness: 0.2, SizePercent: 100, OffsetPercent: 100}
	color := &props.Color{Red: 230, Green: 204, Blue: 255} // Lavender

	return []core.Row{
		row.New(0).WithStyle(&props.Cell{BorderType: border.Top | border.Left | border.Right, BorderThickness: 0.2, BackgroundColor: color}).Add(
			text.NewCol(12, "DEBIT NOTE", props.Text{Style: fontstyle.Bold, Size: 12, Align: align.Center, Bottom: padding + 1, Top: padding}),
		),
		row.New(0).WithStyle(&props.Cell{BorderType: border.Top | border.Left | border.Right, BorderThickness: 0.2}).Add(
			text.NewCol(2, "To", lblTop),
			text.NewCol(4, ": Perkakas Rekadaya Nusantara", valTop),
			line.NewCol(1, sep),
			text.NewCol(2, "Debit Note No", lblTop),
			text.NewCol(3, ": DN/2026/004", valTop),
		),
		row.New(0).WithStyle(&props.Cell{BorderType: border.Left | border.Right, BorderThickness: 0.2}).Add(
			text.NewCol(2, "Attn.", lblMid),
			text.NewCol(4, ": Finance Dept.", valMid),
			line.NewCol(1, sep),
			text.NewCol(2, "Issue Date", lblMid),
			text.NewCol(3, ": 26 March 2026", valMid),
		),
		row.New(0).WithStyle(&props.Cell{BorderType: border.Left | border.Right | border.Bottom, BorderThickness: 0.2}).Add(
			text.NewCol(2, "Reason", lblBot),
			text.NewCol(4, ": Price Adjustment", valBot),
			line.NewCol(1, sep),
			text.NewCol(2, "Orig. Invoice", lblBot),
			text.NewCol(3, ": INV-2026-XYZ", valBot),
		),
	}
}

func (r *DebitNoteRenderer) buildContentSection(items []models.DocumentItem) []core.Row {
	var rows []core.Row
	rows = append(rows, row.New(5))

	hdrCenter := props.Text{Style: fontstyle.Bold, Align: align.Center, Size: 8, Top: 2, Bottom: 2}
	hdrLeft := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}
	valCenter := props.Text{Align: align.Center, Size: 8, Top: 2, Bottom: 2}
	valLeft := props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}

	color := &props.Color{Red: 230, Green: 204, Blue: 255} // Lavender
	hdrCell := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}
	itemCell := &props.Cell{BorderType: border.Full, BorderThickness: 0.2}

	rows = append(rows, row.New(8).Add(
		text.NewCol(1, "#", hdrCenter).WithStyle(hdrCell),
		text.NewCol(6, "Debit Details", hdrLeft).WithStyle(hdrCell),
		text.NewCol(1, "Qty", hdrCenter).WithStyle(hdrCell),
		text.NewCol(2, "Unit Differential", hdrCenter).WithStyle(hdrCell),
		text.NewCol(2, "Debit Amount", hdrCenter).WithStyle(hdrCell),
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
	// unused removed
	summaryValueBold := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}
	emptyCol := props.Text{Align: align.Left, Size: 8, Top: 2, Bottom: 2, Left: 2}

	rows = append(rows, row.New(8).Add(
		text.NewCol(8, "", emptyCol),
		text.NewCol(2, "Total Debit", summaryLabel).WithStyle(summaryCell),
		text.NewCol(2, "Rp 111.000.000", summaryValueBold).WithStyle(summaryCellRight),
	))

	// TERMS
	rows = append(rows, row.New(5))
	outerTop := &props.Cell{BorderType: border.Full, BorderThickness: 0.2, BackgroundColor: color}
	outerBot := &props.Cell{BorderType: border.Left | border.Right | border.Bottom, BorderThickness: 0.2}

	rows = append(rows, row.New(8).WithStyle(outerTop).Add(
		text.NewCol(12, "DEBIT INSTRUCTIONS", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 8, Top: 2, Left: 2}),
	))
	rows = append(rows, row.New(8).WithStyle(outerBot).Add(
		text.NewCol(12, "1. This amount will be debited from your account ledger immediately.", props.Text{Align: align.Left, Size: 8, Top: 2, Left: 2, Bottom: 2}),
	))

	return rows
}
