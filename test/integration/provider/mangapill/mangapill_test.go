//go:build integration

// Package mangapill holds integration tests that scrape the live Mangapill
// site. Run with: go test -tags=integration ./test/integration/...
package mangapill

import (
	"testing"

	"github.com/yukiteruamano/koma/provider/generic"
	"github.com/yukiteruamano/koma/provider/mangapill"
)

func TestSearchAndChaptersIntegration(t *testing.T) {
	scraper := generic.New(mangapill.Config)

	mangas, err := scraper.Search("Death Note")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(mangas) == 0 {
		t.Fatal("expected at least one manga")
	}

	for _, manga := range mangas {
		if manga.Name == "" || manga.URL == "" {
			t.Fatalf("manga has empty name or URL: %+v", manga)
		}
	}

	chapters, err := scraper.ChaptersOf(mangas[0])
	if err != nil {
		t.Fatalf("ChaptersOf failed: %v", err)
	}

	if len(chapters) == 0 {
		t.Fatal("expected at least one chapter")
	}
}
