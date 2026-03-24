package components

import (
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func BuildFooter() core.Row {
	return row.New(25).Add(
		col.New(12).Add(
			text.New("Cozybox", props.Text{Style: fontstyle.Bold, Size: 12}),
			text.New("123 Main St", props.Text{Size: 8, Top: 7}),
			text.New("Anytown, USA 12345", props.Text{Size: 8, Top: 11}),
			text.New("Tel: 123-456-7890 | Email: [EMAIL_ADDRESS]", props.Text{Size: 8, Top: 15}),
		),
	)
}
