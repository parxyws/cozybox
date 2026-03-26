package components

import (
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func BuildHeader() []core.Row {
	return []core.Row{
		// Row 1: Company name
		row.New(10).Add(
			col.New(12).Add(
				text.New("COZYBOX", props.Text{Style: fontstyle.Bold, Size: 24}),
			),
		),
		row.New(2).Add(col.New(12)),
		// Row 2: Line separator
		row.New(1).Add(
			col.New(12).Add(
				line.New(props.Line{SizePercent: 100, Thickness: 0.5}),
			),
		),
		// Row 3: Address details
		row.New(20).Add(
			col.New(12).Add(
				text.New("PT Cozybox Jaya Abadi", props.Text{Size: 10, Style: fontstyle.Bold}),
				text.New("Jl. Raya Bogor KM 25 No. 123", props.Text{Size: 8, Top: 4}),
				text.New("Jakarta, Indonesia 12345", props.Text{Size: 8, Top: 7}),
				text.New("Tel: 123-456-7890 | Email: admin@cozybox.com", props.Text{Size: 8, Top: 10}),
				text.New("www.cozybox.com", props.Text{Size: 8, Top: 13}),
			),
		),

		row.New(15),
	}
}
