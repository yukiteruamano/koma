package zonatmo

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/source"
)

func init() {
	filesystem.SetMemMapFs()
}

func fixtureServer(t *testing.T, path string) *httptest.Server {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)
	return server
}

func TestSearch(t *testing.T) {
	server := fixtureServer(t, "testdata/search.html")
	old := baseURL
	baseURL = server.URL
	defer func() { baseURL = old }()

	s := New()
	mangas, err := s.Search("one piece")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(mangas) == 0 {
		t.Fatal("expected at least one manga")
	}

	for _, m := range mangas {
		if m.Name == "" {
			t.Error("expected a non-empty name")
		}
		if m.URL == "" {
			t.Error("expected a non-empty URL")
		}
		if m.Metadata.Cover.ExtraLarge == "" {
			t.Error("expected a cover URL")
		}
	}

	if mangas[0].Name != "One Piece" {
		t.Errorf("first manga = %q, want One Piece", mangas[0].Name)
	}
}

func TestChaptersOf(t *testing.T) {
	server := fixtureServer(t, "testdata/manga.html")
	old := baseURL
	baseURL = server.URL
	defer func() { baseURL = old }()

	s := New()
	manga := &source.Manga{Name: "One Piece", URL: server.URL + "/library/manga/31322/one-piece", Source: s}

	chapters, err := s.ChaptersOf(manga)
	if err != nil {
		t.Fatalf("ChaptersOf failed: %v", err)
	}

	if len(chapters) == 0 {
		t.Fatal("expected at least one chapter")
	}

	for i, ch := range chapters {
		if ch.Name == "" {
			t.Errorf("chapter %d has an empty name", i)
		}
		if ch.URL == "" {
			t.Errorf("chapter %d has an empty URL", i)
		}
		if ch.Index != uint16(i) {
			t.Errorf("chapter %d index = %d, want %d", i, ch.Index, i)
		}
		if ch.Manga != manga {
			t.Error("chapter Manga back-reference not set")
		}
	}
}

func TestPagesOf(t *testing.T) {
	server := fixtureServer(t, "testdata/reader.html")
	old := baseURL
	baseURL = server.URL
	defer func() { baseURL = old }()

	s := New()
	chapter := &source.Chapter{Name: "Capítulo 1190", URL: server.URL + "/view_uploads/1020781", Manga: &source.Manga{Source: s}}

	pages, err := s.PagesOf(chapter)
	if err != nil {
		t.Fatalf("PagesOf failed: %v", err)
	}

	if len(pages) == 0 {
		t.Fatal("expected at least one page")
	}

	for i, p := range pages {
		if p.URL == "" {
			t.Errorf("page %d has an empty URL", i)
		}
		if p.Extension == "" {
			t.Errorf("page %d has an empty extension", i)
		}
		if p.Index != uint16(i) {
			t.Errorf("page %d index = %d, want %d", i, p.Index, i)
		}
	}
}

func TestAbsoluteURL(t *testing.T) {
	tests := []struct {
		name string
		href string
		want string
	}{
		{name: "absolute", href: "https://zonatmo.org/library/manga/31322/x", want: "https://zonatmo.org/library/manga/31322/x"},
		{name: "protocol-relative", href: "//zonatmo.org/library/x", want: "https://zonatmo.org/library/x"},
		{name: "empty", href: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := absoluteURL(tt.href); got != tt.want {
				t.Errorf("absoluteURL(%q) = %q, want %q", tt.href, got, tt.want)
			}
		})
	}
}

func TestURLID(t *testing.T) {
	tests := []struct {
		name string
		href string
		want string
	}{
		{name: "view upload", href: "https://zonatmo.org/view_uploads/1020781", want: "1020781"},
		{name: "library slug", href: "https://zonatmo.org/library/manga/31322/one-piece", want: "one-piece"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := urlID(tt.href); got != tt.want {
				t.Errorf("urlID(%q) = %q, want %q", tt.href, got, tt.want)
			}
		})
	}
}

func TestSearchCacheHitRehydratesSource(t *testing.T) {
	filesystem.SetMemMapFs()

	first := New()
	manga := &source.Manga{Name: "Cache Manga", URL: "https://zonatmo.org/library/manga/99999/cache-manga"}
	if err := first.cache.mangas.Set("cache-query", []*source.Manga{manga}); err != nil {
		t.Fatalf("cache Set failed: %v", err)
	}

	// a fresh instance reloads from disk, losing the Manga.Source back-reference
	second := New()
	mangas, err := second.Search("cache-query")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(mangas) != 1 {
		t.Fatalf("got %d mangas, want 1", len(mangas))
	}
	if mangas[0].Source != second {
		t.Error("Manga.Source was not rehydrated from the cache")
	}
}

func TestChaptersCacheHitRehydratesManga(t *testing.T) {
	filesystem.SetMemMapFs()

	first := New()
	manga := &source.Manga{Name: "Cache Manga", ID: "cache-manga-id", URL: "https://zonatmo.org/library/manga/99999/cache-manga"}
	chapters := []*source.Chapter{
		{Name: "Capítulo 1", Manga: manga},
		{Name: "Capítulo 2", Manga: manga},
	}
	if err := first.cache.chapters.Set(manga.ID, chapters); err != nil {
		t.Fatalf("cache Set failed: %v", err)
	}

	// a fresh instance reloads from disk, losing the Chapter.Manga back-reference
	second := New()
	got, err := second.ChaptersOf(manga)
	if err != nil {
		t.Fatalf("ChaptersOf failed: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d chapters, want 2", len(got))
	}
	for i, chapter := range got {
		if chapter.Manga != manga {
			t.Errorf("chapter %d Manga was not rehydrated from the cache", i)
		}
	}
}
