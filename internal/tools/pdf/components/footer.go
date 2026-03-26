package components

import (
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func BuildFooter() core.Row {
	// row.New(height) gives the footer a total height.
	// You can increase '15' if you want a taller logo.
	return row.New(15).Add(
		// Create an empty column on the left spanning 9/12 grids
		col.New(10),

		// Create an image column on the right spanning 3/12 grids
		image.NewFromFileCol(2, "../../../../cozybox.png", props.Rect{
			Center:  true,
			Percent: 100, // Adjust this (1 to 100) for a custom relative size inside the column
		}),
	)
}
