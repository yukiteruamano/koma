package inline

import (
	"testing"

	"github.com/yukiteruamano/koma/source"
)

func TestParseMangaPicker(t *testing.T) {
	mangas := []*source.Manga{
		{Name: "a"},
		{Name: "b"},
		{Name: "c"},
	}

	tests := []struct {
		name        string
		description string
		query       string
		wantName    string
		wantErr     bool
	}{
		{name: "first", description: "first", wantName: "a"},
		{name: "last", description: "last", wantName: "c"},
		{name: "exact match", description: "exact", wantName: "b", query: "b"},
		{name: "index 1", description: "1", wantName: "b"},
		{name: "large index clamps", description: "999999", wantName: "c"},
		{name: "invalid pattern", description: "bogus", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			picker, err := ParseMangaPicker(tt.query, tt.description)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := picker(mangas)
			if got == nil {
				t.Fatal("expected a picked manga")
			}
			if got.Name != tt.wantName {
				t.Errorf("picked %q, want %q", got.Name, tt.wantName)
			}
		})
	}
}

func TestParseChaptersFilterBounds(t *testing.T) {
	chapters := []*source.Chapter{
		{Name: "c1", Index: 1},
		{Name: "c2", Index: 2},
	}

	tests := []struct {
		name        string
		description string
		wantNames   []string
	}{
		{name: "single out of range", description: "999", wantNames: []string{}},
		{name: "range beyond end", description: "2-5", wantNames: []string{"c2"}},
		{name: "first on non-empty", description: "first", wantNames: []string{"c1"}},
		{name: "last on non-empty", description: "last", wantNames: []string{"c2"}},
		{name: "exact index", description: "1", wantNames: []string{"c2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := ParseChaptersFilter(tt.description)
			if err != nil {
				t.Fatalf("ParseChaptersFilter(%q) failed: %v", tt.description, err)
			}

			// must not panic on out-of-range indices
			got, err := filter(chapters)
			if err != nil {
				t.Fatalf("filter failed: %v", err)
			}

			if len(got) != len(tt.wantNames) {
				t.Fatalf("got %d chapters (%v), want %v", len(got), names(got), tt.wantNames)
			}
			for i, want := range tt.wantNames {
				if got[i].Name != want {
					t.Errorf("chapter %d = %q, want %q", i, got[i].Name, want)
				}
			}
		})
	}
}

func TestParseChaptersFilterEmpty(t *testing.T) {
	filter, err := ParseChaptersFilter("first")
	if err != nil {
		t.Fatal(err)
	}

	got, err := filter(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected no chapters for empty input, got %v", names(got))
	}
}

func names(chapters []*source.Chapter) []string {
	out := make([]string, len(chapters))
	for i, c := range chapters {
		out[i] = c.Name
	}
	return out
}

func TestParseChaptersFilter(t *testing.T) {
	chapters := []*source.Chapter{
		{Name: "Chapter 1", Index: 1},
		{Name: "Chapter 2", Index: 2},
		{Name: "Chapter 3", Index: 3},
		{Name: "Chapter 4", Index: 4},
	}

	tests := []struct {
		name        string
		description string
		wantCount   int
		wantErr     bool
	}{
		{name: "first", description: "first", wantCount: 1},
		{name: "last", description: "last", wantCount: 1},
		{name: "all", description: "all", wantCount: 4},
		{name: "range 2-3", description: "2-3", wantCount: 2},
		{name: "single 1", description: "1", wantCount: 1},
		{name: "large range clamps", description: "1-999999", wantCount: 3},
		{name: "substring", description: "@Chapter 2@", wantCount: 1},
		{name: "invalid", description: "nonsense", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := ParseChaptersFilter(tt.description)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got, err := filter(chapters)
			if err != nil {
				t.Fatalf("filter failed: %v", err)
			}

			if len(got) != tt.wantCount {
				t.Errorf("filtered %d chapters, want %d", len(got), tt.wantCount)
			}
		})
	}
}
