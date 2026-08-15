package integration

import (
	"github.com/yukiteruamano/koma/integration/anilist"
	"github.com/yukiteruamano/koma/source"
)

// Integrator is the interface that wraps the basic integration methods.
type Integrator interface {
	// MarkRead marks a chapter as read
	MarkRead(chapter *source.Chapter) error
}

var (
	Anilist Integrator = anilist.New()
)
