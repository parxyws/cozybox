package pdf

import (
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core/entity"
)

func BuildConfig() *entity.Config {
	return config.NewBuilder().WithPageSize(pagesize.A4).WithOrientation(orientation.Vertical).Build()
}
