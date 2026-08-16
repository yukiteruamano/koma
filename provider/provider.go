package provider

import (
	"github.com/yukiteruamano/koma/source"
)

type Provider struct {
	ID           string
	Name         string
	CreateSource func() (source.Source, error)
}

func (p Provider) String() string {
	return p.Name
}

func Builtins() []*Provider {
	return builtinProviders
}

func Get(name string) (*Provider, bool) {
	for _, provider := range Builtins() {
		if provider.Name == name {
			return provider, true
		}
	}

	return nil, false
}
