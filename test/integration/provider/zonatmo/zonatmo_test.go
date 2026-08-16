//go:build integration

// Package zonatmo holds integration tests against the live Zonatmo site.
// Run with: go test -tags=integration ./test/integration/...
package zonatmo

import (
	"testing"

	"github.com/yukiteruamano/koma/provider/zonatmo"
)

func TestSearchIntegration(t *testing.T) {
	src := zonatmo.New()

	mangas, err := src.Search("one piece")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(mangas) == 0 {
		t.Fatal("expected at least one result")
	}

	if mangas[0].Name == "" || mangas[0].URL == "" {
		t.Fatalf("manga missing name/url: %+v", mangas[0])
	}
}

func TestChaptersIntegration(t *testing.T) {
	src := zonatmo.New()

	mangas, err := src.Search("one piece")
	if err != nil || len(mangas) == 0 {
		t.Fatalf("Search failed: %v", err)
	}

	chapters, err := src.ChaptersOf(mangas[0])
	if err != nil {
		t.Fatalf("ChaptersOf failed: %v", err)
	}

	if len(chapters) == 0 {
		t.Fatal("expected at least one chapter")
	}
}

func TestPagesIntegration(t *testing.T) {
	src := zonatmo.New()

	mangas, err := src.Search("one piece")
	if err != nil || len(mangas) == 0 {
		t.Fatalf("Search failed: %v", err)
	}

	chapters, err := src.ChaptersOf(mangas[0])
	if err != nil || len(chapters) == 0 {
		t.Fatalf("ChaptersOf failed: %v", err)
	}

	pages, err := src.PagesOf(chapters[0])
	if err != nil {
		t.Fatalf("PagesOf failed: %v", err)
	}

	if len(pages) == 0 {
		t.Fatal("expected at least one page")
	}

	if pages[0].URL == "" {
		t.Fatal("expected a page URL")
	}
}
