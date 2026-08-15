package source

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/samber/lo"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/constant"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/util"
)

func init() {
	filesystem.SetMemMapFs()
	viper.Set(key.FormatsUse, constant.FormatPDF)
}

// setTestConfig pins the viper keys the source tests depend on, so results do
// not depend on the order other tests mutated the shared viper state.
func setTestConfig(format string) {
	viper.Set(key.DownloaderPath, "/downloads")
	viper.Set(key.DownloaderChapterNameTemplate, "{chapter}")
	viper.Set(key.DownloaderCreateMangaDir, true)
	viper.Set(key.DownloaderCreateVolumeDir, false)
	viper.Set(key.DownloaderEscapeWhitespace, true)
	viper.Set(key.FormatsUse, format)
}

var testChapter = Chapter{
	Name:   "test chapter",
	URL:    "https://example.com",
	Index:  1,
	ID:     "test",
	Pages:  []*Page{},
	Manga:  &testManga,
	Volume: "1",
}

func TestChapter_Filename(t *testing.T) {
	setTestConfig(constant.FormatPDF)
	Convey("Given a chapter", t, func() {
		Convey("When Filename is called", func() {
			Convey("It should return a sanitized filename", func() {
				const template = "&{index}! {chapter}// {volume} 28922@ {manga}"
				viper.Set(key.DownloaderChapterNameTemplate, template)
				filename := testChapter.Filename()

				Convey("It should match the given template", func() {
					So(filename, ShouldEqual, util.SanitizeFilename(fmt.Sprintf("&%d! %s// %s 28922@ %s.pdf", testChapter.Index, testChapter.Name, testChapter.Volume, testChapter.Manga.Name)))
				})
			})
		})
	})
}

func TestChapter_IsDownloaded(t *testing.T) {
	setTestConfig(constant.FormatCBZ)
	viper.Set(key.DownloaderPath, "/downloads")
	viper.Set(key.DownloaderChapterNameTemplate, "{chapter}")
	viper.Set(key.DownloaderCreateMangaDir, true)
	viper.Set(key.FormatsUse, constant.FormatCBZ)

	newChapter := func() *Chapter {
		manga := &Manga{Name: "Manga A"}
		return &Chapter{Name: "Chapter 1", Index: 1, Manga: manga}
	}

	t.Run("missing file reports not downloaded", func(t *testing.T) {
		chapter := newChapter()
		if chapter.IsDownloaded() {
			t.Error("expected not downloaded when the file is missing")
		}
	})

	t.Run("existing file reports downloaded", func(t *testing.T) {
		chapter := newChapter()
		path := lo.Must(chapter.Path(false))
		if err := filesystem.Api().WriteFile(path, []byte("data"), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		if !chapter.IsDownloaded() {
			t.Error("expected downloaded after the file exists")
		}
	})

	t.Run("cached flag prevents re-stat", func(t *testing.T) {
		chapter := newChapter()
		path := lo.Must(chapter.Path(false))
		if err := filesystem.Api().WriteFile(path, []byte("data"), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		if !chapter.IsDownloaded() {
			t.Fatal("setup: expected downloaded")
		}
		if err := filesystem.Api().Remove(path); err != nil {
			t.Fatal(err)
		}
		if !chapter.IsDownloaded() {
			t.Error("expected the cached flag to still report downloaded")
		}
	})
}

func TestChapter_PathVolumeDir(t *testing.T) {
	setTestConfig(constant.FormatCBZ)
	filesystem.SetMemMapFs()
	viper.Set(key.DownloaderPath, "/downloads")
	viper.Set(key.DownloaderChapterNameTemplate, "{chapter}")
	viper.Set(key.DownloaderCreateMangaDir, true)
	viper.Set(key.DownloaderCreateVolumeDir, true)
	viper.Set(key.FormatsUse, constant.FormatCBZ)

	manga := &Manga{Name: "Manga V"}
	chapter := &Chapter{Name: "Ch", Index: 1, Volume: "Vol.2", Manga: manga}

	path := lo.Must(chapter.Path(false))
	want := filepath.Join("/downloads", "Manga_V", "Vol.2", "Ch.cbz")
	if path != want {
		t.Errorf("Path() = %q, want %q", path, want)
	}
}

func TestChapterFormattedName(t *testing.T) {
	setTestConfig(constant.FormatPDF)
	viper.Set(key.DownloaderEscapeWhitespace, true)

	manga := &Manga{
		Name:   "Test Manga",
		Source: fakeSource{},
		Chapters: []*Chapter{
			{Name: "c1", Index: 1},
			{Name: "c2", Index: 2},
			{Name: "c3", Index: 3},
		},
	}

	chapter := &Chapter{
		Name:   "The Chapter",
		Index:  7,
		Volume: "Vol.1",
		Manga:  manga,
	}

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{name: "manga", template: "{manga}", want: "Test Manga"},
		{name: "chapter", template: "{chapter}", want: "The Chapter"},
		{name: "index", template: "{index}", want: "7"},
		{name: "padded index", template: "{padded-index}", want: "0007"},
		{name: "chapters count", template: "{chapters-count}", want: "3"},
		{name: "volume", template: "{volume}", want: "Vol.1"},
		{name: "source", template: "{source}", want: "fake"},
		{name: "combined", template: "[{padded-index}] {chapter} ({manga})", want: "[0007] The Chapter (Test Manga)"},
		{name: "no placeholders", template: "fixed name", want: "fixed name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set(key.DownloaderChapterNameTemplate, tt.template)
			if got := chapter.formattedName(); got != tt.want {
				t.Errorf("formattedName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChapterComicInfoFields(t *testing.T) {
	setTestConfig(constant.FormatPDF)
	viper.Set(key.MetadataComicInfoXMLAddDate, false)

	manga := &Manga{Name: "Series Name"}
	manga.Metadata.Genres = []string{"Action", "Adventure"}
	manga.Metadata.Tags = []string{"tag1", "tag2"}
	manga.Metadata.Staff.Story = []string{"Writer A"}
	manga.Metadata.Staff.Art = []string{"Artist B"}
	manga.Metadata.Staff.Translation = []string{"TL Group"}
	manga.Metadata.Staff.Lettering = []string{"Lett Group"}

	chapter := &Chapter{
		Name:  "Chapter Name",
		Index: 3,
		URL:   "https://example.com/ch/3",
		Manga: manga,
		Pages: []*Page{{Index: 0}, {Index: 1}},
	}

	info := chapter.ComicInfo()

	if info.Title != "Chapter Name" {
		t.Errorf("Title = %q, want Chapter Name", info.Title)
	}
	if info.Series != "Series Name" {
		t.Errorf("Series = %q, want Series Name", info.Series)
	}
	if info.Number != 3 {
		t.Errorf("Number = %d, want 3", info.Number)
	}
	if info.Genre != "Action,Adventure" {
		t.Errorf("Genre = %q, want Action,Adventure", info.Genre)
	}
	if info.Writer != "Writer A" {
		t.Errorf("Writer = %q, want Writer A", info.Writer)
	}
	if info.Penciller != "Artist B" {
		t.Errorf("Penciller = %q, want Artist B", info.Penciller)
	}
	if info.Translator != "TL Group" {
		t.Errorf("Translator = %q, want TL Group", info.Translator)
	}
	if info.PageCount != 2 {
		t.Errorf("PageCount = %d, want 2", info.PageCount)
	}
}
