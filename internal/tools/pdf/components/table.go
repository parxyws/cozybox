package components

import (
	"fmt"

	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/parxyws/cozybox/internal/models"
)

func BuildTableList(items []models.DocumentItem) []core.Row {
	var rows []core.Row

	rows = append(rows, row.New(8).Add(
		col.New(1).Add(text.New("#", props.Text{Style: fontstyle.Bold, Size: 8})),
		col.New(4).Add(text.New("Description", props.Text{Style: fontstyle.Bold, Size: 8})),
		col.New(1).Add(text.New("Qty", props.Text{Style: fontstyle.Bold, Size: 8})),
		col.New(2).Add(text.New("Unit Price", props.Text{Style: fontstyle.Bold, Size: 8})),
		col.New(1).Add(text.New("Disc%", props.Text{Style: fontstyle.Bold, Size: 8})),
		col.New(1).Add(text.New("Tax%", props.Text{Style: fontstyle.Bold, Size: 8})),
		col.New(2).Add(text.New("Amount", props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Right})),
	))

	for i, item := range items {
		rows = append(rows, row.New(7).Add(
			col.New(1).Add(text.New(fmt.Sprintf("%d", i+1), props.Text{Size: 8})),
			col.New(4).Add(text.New(item.Description, props.Text{Size: 8})),
			col.New(1).Add(text.New(item.Quantity.String(), props.Text{Size: 8})),
			col.New(2).Add(text.New(item.UnitPrice.StringFixed(2), props.Text{Size: 8})),
			col.New(1).Add(text.New(item.DiscountPct.String()+"%", props.Text{Size: 8})),
			col.New(1).Add(text.New(item.TaxPct.String()+"%", props.Text{Size: 8})),
			col.New(2).Add(text.New(item.Amount.StringFixed(2), props.Text{Size: 8, Align: align.Right})),
		))
	}
	return rows

}
