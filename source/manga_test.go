package source

import (
	"encoding/json"
	"testing"

	"github.com/samber/mo"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/anilist"
	"github.com/yukiteruamano/koma/constant"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/util"
)

func init() {
	filesystem.SetMemMapFs()
}

type testSource struct{}

func (t testSource) Name() string {
	return "test"
}

func (t testSource) ID() string {
	return "test"
}

func (t testSource) Search(string) (mangas []*Manga, err error) {
	return
}

func (t testSource) ChaptersOf(*Manga) (chapters []*Chapter, err error) {
	return
}

func (t testSource) PagesOf(*Chapter) (pages []*Page, err error) {
	return
}

var testManga = Manga{
	Name:     "Death Note",
	URL:      "https://example.com",
	Index:    1,
	ID:       "test",
	Chapters: []*Chapter{},
	Source:   &testSource{},
}

func TestManga_Filename(t *testing.T) {
	setTestConfig(constant.FormatPDF)
	Convey("Given a manga", t, func() {
		Convey("When Filename is called", func() {
			Convey("It should return a sanitized filename", func() {
				So(testManga.Dirname(), ShouldEqual, util.SanitizeFilename(testManga.Name))
			})
		})
	})
}

func TestManga_Path(t *testing.T) {
	setTestConfig(constant.FormatPDF)

	newManga := func() *Manga {
		return &Manga{Name: "Path Manga", Source: fakeSource{}}
	}

	t.Run("non-temp path is a directory", func(t *testing.T) {
		path, err := newManga().Path(false)
		if err != nil {
			t.Fatal(err)
		}
		if path == "" {
			t.Error("expected a non-empty path")
		}
		isDir, err := filesystem.Api().IsDir(path)
		if err != nil {
			t.Fatal(err)
		}
		if !isDir {
			t.Errorf("expected %q to be a directory", path)
		}
	})

	t.Run("temp path is a directory", func(t *testing.T) {
		path, err := newManga().Path(true)
		if err != nil {
			t.Fatal(err)
		}
		if path == "" {
			t.Error("expected a non-empty path")
		}
		isDir, err := filesystem.Api().IsDir(path)
		if err != nil {
			t.Fatal(err)
		}
		if !isDir {
			t.Errorf("expected %q to be a directory", path)
		}
	})
}

func TestManga_PopulateMetadata(t *testing.T) {
	viper.Set(key.MetadataComicInfoXMLTagRelevanceThreshold, 70)

	al := testAnilistManga()
	m := &Manga{Name: "Test Series"}
	m.Anilist = mo.Some(al)

	err := m.PopulateMetadata(func(string) {})
	if err != nil {
		t.Fatalf("PopulateMetadata failed: %v", err)
	}

	if len(m.Metadata.Genres) != 1 || m.Metadata.Genres[0] != "Action" {
		t.Errorf("genres = %v, want [Action]", m.Metadata.Genres)
	}
	if m.Metadata.Summary != "hello\nworld bold" {
		t.Errorf("summary = %q, want %q", m.Metadata.Summary, "hello\nworld bold")
	}
	if len(m.Metadata.Tags) != 1 || m.Metadata.Tags[0] != "High" {
		t.Errorf("tags = %v, want [High] (rank >= threshold)", m.Metadata.Tags)
	}
	if len(m.Metadata.Staff.Story) != 1 || m.Metadata.Staff.Story[0] != "Writer B" {
		t.Errorf("story = %v, want [Writer B]", m.Metadata.Staff.Story)
	}
	if len(m.Metadata.Staff.Art) != 1 || m.Metadata.Staff.Art[0] != "Artist C" {
		t.Errorf("art = %v, want [Artist C]", m.Metadata.Staff.Art)
	}
	if m.Metadata.Cover.ExtraLarge != "https://cover" {
		t.Errorf("cover = %q, want https://cover", m.Metadata.Cover.ExtraLarge)
	}
	if m.Metadata.Status != "RELEASING" {
		t.Errorf("status = %q, want RELEASING", m.Metadata.Status)
	}
	if !m.populated {
		t.Error("populated should be true after a successful fetch")
	}
}

func TestManga_PopulateMetadataIdempotent(t *testing.T) {
	m := &Manga{Name: "Test Series"}
	m.Anilist = mo.Some(testAnilistManga())

	if err := m.PopulateMetadata(func(string) {}); err != nil {
		t.Fatalf("first populate failed: %v", err)
	}

	// populate again: should short-circuit, not re-fetch or error
	if err := m.PopulateMetadata(func(string) {}); err != nil {
		t.Fatalf("second populate failed: %v", err)
	}
}

func testAnilistManga() *anilist.Manga {
	data := map[string]any{
		"title":       map[string]any{"english": "Test Series"},
		"description": "hello<br>world <b>bold</b>",
		"genres":      []string{"Action"},
		"tags": []map[string]any{
			{"name": "High", "rank": 90},
			{"name": "Low", "rank": 10},
		},
		"characters": map[string]any{
			"nodes": []map[string]any{{"name": map[string]any{"full": "Char A"}}},
		},
		"staff": map[string]any{
			"edges": []map[string]any{
				{"role": "Story", "node": map[string]any{"name": map[string]any{"full": "Writer B"}}},
				{"role": "Art", "node": map[string]any{"name": map[string]any{"full": "Artist C"}}},
			},
		},
		"coverImage": map[string]any{"extraLarge": "https://cover"},
		"status":     "RELEASING",
		"siteUrl":    "https://anilist.co/anime/1",
		"idMal":      42,
	}

	buf, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	al := &anilist.Manga{}
	if err := json.Unmarshal(buf, al); err != nil {
		panic(err)
	}
	return al
}

func TestManga_GetCover(t *testing.T) {
	tests := []struct {
		name string
		xl   string
		lg   string
		md   string
		want string
	}{
		{name: "extra large first", xl: "xl", lg: "lg", md: "md", want: "xl"},
		{name: "falls back to large", lg: "lg", md: "md", want: "lg"},
		{name: "falls back to medium", md: "md", want: "md"},
		{name: "no cover", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manga{Name: "m"}
			m.Metadata.Cover.ExtraLarge = tt.xl
			m.Metadata.Cover.Large = tt.lg
			m.Metadata.Cover.Medium = tt.md

			got, err := m.GetCover()
			if tt.want == "" {
				if err == nil {
					t.Fatal("expected an error when no cover is available")
				}
				return
			}
			if err != nil {
				t.Fatalf("GetCover failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("GetCover = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestManga_SeriesJSON(t *testing.T) {
	Convey("Given a manga", t, func() {
		Convey("When SeriesJSON is called", func() {
			buf := testManga.SeriesJSON()
			Convey("It should return a json buffer", func() {
				So(buf, ShouldNotBeEmpty)
			})
		})
	})
}
