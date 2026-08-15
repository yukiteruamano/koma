package downloader

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/constant"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/test/testutil"
)

func setupDownload(t *testing.T, format string) {
	t.Helper()

	filesystem.SetMemMapFs()
	viper.Set(key.DownloaderPath, "/downloads")
	viper.Set(key.DownloaderChapterNameTemplate, "{chapter}")
	viper.Set(key.DownloaderAsync, false)
	viper.Set(key.DownloaderCreateMangaDir, true)
	viper.Set(key.DownloaderCreateVolumeDir, false)
	viper.Set(key.DownloaderRedownloadExisting, false)
	viper.Set(key.HistorySaveOnDownload, false)
	viper.Set(key.MetadataFetchAnilist, false)
	viper.Set(key.MetadataSeriesJSON, false)
	viper.Set(key.DownloaderDownloadCover, false)
	viper.Set(key.DownloaderEscapeWhitespace, true)
	viper.Set(key.FormatsUse, format)
}

func TestDownloadPlainStreamsPages(t *testing.T) {
	setupDownload(t, constant.FormatPlain)

	chapter, _, _ := testutil.NewChapter(t, "Plain Manga", 3)

	path, err := Download(chapter, func(string) {})
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if path != filepath.Join("/downloads", "Plain_Manga", "Plain_Manga") {
		t.Errorf("unexpected plain dir path: %q", path)
	}

	entries, err := filesystem.Api().ReadDir(path)
	if err != nil {
		t.Fatalf("ReadDir(%q) failed: %v", path, err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 page files, got %d", len(entries))
	}

	// streaming must not keep page contents in memory
	for _, page := range chapter.Pages {
		if page.Contents != nil {
			t.Error("plain download should not buffer page contents in memory")
		}
		if page.Size == 0 {
			t.Error("streamed page should record its size")
		}
	}
}

func TestDownloadCbz(t *testing.T) {
	setupDownload(t, constant.FormatCBZ)

	chapter, _, _ := testutil.NewChapter(t, "Cbz Manga", 2)

	path, err := Download(chapter, func(string) {})
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if filepath.Ext(path) != ".cbz" {
		t.Errorf("expected .cbz output, got %q", path)
	}

	if _, err := filesystem.Api().Stat(path); err != nil {
		t.Errorf("expected cbz file to exist: %v", err)
	}
}

func TestDownloadSkipsAlreadyDownloaded(t *testing.T) {
	setupDownload(t, constant.FormatPlain)

	chapter, _, _ := testutil.NewChapter(t, "Skip Manga", 2)

	first, err := Download(chapter, func(string) {})
	if err != nil {
		t.Fatalf("first Download failed: %v", err)
	}

	second, err := Download(chapter, func(string) {})
	if err != nil {
		t.Fatalf("second Download failed: %v", err)
	}

	if first != second {
		t.Errorf("skipped download should return the same path: %q vs %q", first, second)
	}
}

func TestDownloadRedownloadExisting(t *testing.T) {
	setupDownload(t, constant.FormatPlain)
	viper.Set(key.DownloaderRedownloadExisting, true)

	chapter, _, _ := testutil.NewChapter(t, "Redownload Manga", 2)

	if _, err := Download(chapter, func(string) {}); err != nil {
		t.Fatalf("first Download failed: %v", err)
	}

	if _, err := Download(chapter, func(string) {}); err != nil {
		t.Fatalf("redownload failed: %v", err)
	}
}

func TestDownloadMissingPages(t *testing.T) {
	setupDownload(t, constant.FormatPlain)

	chapter, fake, _ := testutil.NewChapter(t, "Empty Manga", 0)
	fake.Pages = chapter.Pages // zero pages

	if _, err := Download(chapter, func(string) {}); err == nil {
		t.Fatal("expected an error for a chapter without pages")
	}
}
