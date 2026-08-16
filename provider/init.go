package provider

import (
	"github.com/yukiteruamano/koma/provider/generic"
	"github.com/yukiteruamano/koma/provider/mangadex"
	"github.com/yukiteruamano/koma/provider/mangapill"
	"github.com/yukiteruamano/koma/provider/weebcentral"
	"github.com/yukiteruamano/koma/source"
)

var builtinProviders = []*Provider{
	{
		ID:   mangadex.ID,
		Name: mangadex.Name,
		CreateSource: func() (source.Source, error) {
			return mangadex.New(), nil
		},
	},
	{
		ID:   weebcentral.ID,
		Name: weebcentral.Name,
		CreateSource: func() (source.Source, error) {
			return weebcentral.New(), nil
		},
	},
}

func init() {
	for _, conf := range []*generic.Configuration{
		mangapill.Config,
	} {
		conf := conf
		builtinProviders = append(builtinProviders, &Provider{
			ID:   conf.ID(),
			Name: conf.Name,
			CreateSource: func() (source.Source, error) {
				return generic.New(conf), nil
			},
		})
	}
}
