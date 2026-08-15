package mangadex

import (
	"testing"

	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/source"
)

func TestPageURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		hash    string
		page    string
		want    string
	}{
		{
			name:    "normal",
			baseURL: "https://uploads.mangadex.org",
			hash:    "abc123",
			page:    "1-2.png",
			want:    "https://uploads.mangadex.org/data/abc123/1-2.png",
		},
		{
			name:    "empty hash",
			baseURL: "https://example.com",
			hash:    "",
			page:    "x.jpg",
			want:    "https://example.com/data//x.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pageURL(tt.baseURL, tt.hash, tt.page); got != tt.want {
				t.Errorf("pageURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCacherSetGet(t *testing.T) {
	filesystem.SetMemMapFs()

	c := newCacher[[]*source.Chapter]("test_chapters")
	chapters := []*source.Chapter{{Name: "ch1", Index: 1}}

	if err := c.Set("manga-1", chapters); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, ok := c.Get("manga-1").Get()
	if !ok {
		t.Fatal("expected a cache hit")
	}

	if len(got) != 1 || got[0].Name != "ch1" {
		t.Errorf("unexpected cached value: %+v", got)
	}

	if c.Get("missing").IsPresent() {
		t.Error("expected a cache miss")
	}
}
