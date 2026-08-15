//go:build integration

// Package mangadex holds integration tests that hit the live MangaDex API.
// Run with: go test -tags=integration ./test/integration/...
package mangadex

import (
	"testing"

	"github.com/yukiteruamano/koma/provider/mangadex"
)

func TestSearchIntegration(t *testing.T) {
	dex := mangadex.New()

	mangas, err := dex.Search("Death Note")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(mangas) == 0 {
		t.Fatal("expected at least one result")
	}

	for _, manga := range mangas {
		if manga.Name == "" || manga.URL == "" || manga.ID == "" {
			t.Fatalf("manga missing name/url/id: %+v", manga)
		}
	}
}
