package pdf

import (
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/core"
)

func GetPageHeader() core.Row {
	return row.New(20)
}

func GetPageFooter() core.Row {
	return row.New(20)
}
