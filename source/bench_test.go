package source

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/key"
)

type fakeSource struct{}

func (fakeSource) Name() string                    { return "fake" }
func (fakeSource) ID() string                      { return "fake" }
func (fakeSource) Search(string) ([]*Manga, error) { return nil, nil }
func (fakeSource) ChaptersOf(*Manga) ([]*Chapter, error) {
	return nil, nil
}
func (fakeSource) PagesOf(*Chapter) ([]*Page, error) { return nil, nil }

func BenchmarkFormattedName(b *testing.B) {
	viper.SetDefault(key.DownloaderChapterNameTemplate, "[{padded-index}] {chapter}")
	viper.SetDefault(key.DownloaderEscapeWhitespace, true)
	viper.SetDefault(key.FormatsUse, "cbz")

	p := fakeSource{}
	chapter := &Chapter{
		Name:   "The Awkward ? Title",
		Index:  1,
		Volume: "Volume 01",
		Manga: &Manga{
			Name:   "Example Manga: Chapter",
			Source: p,
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chapter.Filename()
	}
}
