package source

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
)

func TestDownloadPagesConcurrent(t *testing.T) {
	viper.Set(key.DownloaderAsync, true)
	viper.Set(key.DownloaderConcurrency, 4)

	const count = 32
	chapter := &Chapter{
		Name:  "concurrent test",
		Index: 1,
		Pages: make([]*Page, count),
	}

	for i := 0; i < count; i++ {
		chapter.Pages[i] = &Page{
			Index:    uint16(i),
			Contents: bytes.NewBuffer(make([]byte, 1024)),
			Size:     1024,
		}
	}

	// URL is empty and Contents are set, so Page.Download returns without
	// performing any network request.
	err := chapter.DownloadPages(false, func(string) {})
	if err != nil {
		t.Fatalf("DownloadPages returned error: %v", err)
	}

	if chapter.size != count*1024 {
		t.Fatalf("size mismatch: got %d, want %d", chapter.size, count*1024)
	}

	if downloaded, ok := chapter.isDownloaded.Get(); !ok || !downloaded {
		t.Fatalf("chapter should be marked as downloaded, got %+v", chapter.isDownloaded)
	}
}

func TestDownloadPagesSequential(t *testing.T) {
	viper.Set(key.DownloaderAsync, false)

	const count = 8
	chapter := &Chapter{
		Name:  "sequential test",
		Index: 1,
		Pages: make([]*Page, count),
	}

	for i := 0; i < count; i++ {
		chapter.Pages[i] = &Page{
			Index:    uint16(i),
			Contents: bytes.NewBuffer(make([]byte, 512)),
			Size:     512,
		}
	}

	err := chapter.DownloadPages(false, func(string) {})
	if err != nil {
		t.Fatalf("DownloadPages returned error: %v", err)
	}

	if chapter.size != count*512 {
		t.Fatalf("size mismatch: got %d, want %d", chapter.size, count*512)
	}
}

func TestDownloadPagesNilPage(t *testing.T) {
	viper.Set(key.DownloaderAsync, true)
	viper.Set(key.DownloaderConcurrency, 4)

	chapter := &Chapter{
		Name:  "nil page test",
		Index: 1,
		Pages: []*Page{
			{Index: 0, Contents: bytes.NewBuffer(make([]byte, 1)), Size: 1},
			nil,
		},
	}

	err := chapter.DownloadPages(false, func(string) {})
	if err == nil {
		t.Fatal("expected error for nil page, got nil")
	}
}

func TestDownloadPagesAsyncError(t *testing.T) {
	viper.Set(key.DownloaderAsync, true)
	viper.Set(key.DownloaderConcurrency, 2)

	chapter := &Chapter{
		Name:  "async error",
		Index: 1,
		Pages: []*Page{
			{Index: 0, Contents: bytes.NewBuffer(make([]byte, 1)), Size: 1},
			// no URL and no contents -> Download returns an error
			{Index: 1},
			{Index: 2, Contents: bytes.NewBuffer(make([]byte, 1)), Size: 1},
		},
	}

	err := chapter.DownloadPages(false, func(string) {})
	if err == nil {
		t.Fatal("expected an error when a page fails to download")
	}

	if downloaded, ok := chapter.isDownloaded.Get(); !ok || downloaded {
		t.Errorf("expected isDownloaded to be false after an error, got %+v", chapter.isDownloaded)
	}
}

func TestDownloadPagesTempNotMarkedDownloaded(t *testing.T) {
	viper.Set(key.DownloaderAsync, false)

	chapter := &Chapter{
		Name:  "temp",
		Index: 1,
		Pages: []*Page{
			{Index: 0, Contents: bytes.NewBuffer(make([]byte, 1)), Size: 1},
		},
	}

	if err := chapter.DownloadPages(true, func(string) {}); err != nil {
		t.Fatal(err)
	}

	if downloaded, ok := chapter.isDownloaded.Get(); !ok || downloaded {
		t.Errorf("temp downloads must not mark the chapter as downloaded, got %+v", chapter.isDownloaded)
	}
}

func TestDownloadPagesConcurrencyClamped(t *testing.T) {
	viper.Set(key.DownloaderAsync, true)
	viper.Set(key.DownloaderConcurrency, 0) // invalid -> clamped to 1

	chapter := &Chapter{
		Name:  "clamped",
		Index: 1,
		Pages: []*Page{
			{Index: 0, Contents: bytes.NewBuffer(make([]byte, 1)), Size: 1},
			{Index: 1, Contents: bytes.NewBuffer(make([]byte, 1)), Size: 1},
		},
	}

	if err := chapter.DownloadPages(false, func(string) {}); err != nil {
		t.Fatalf("DownloadPages failed: %v", err)
	}
}

func TestDownloadPagesToDir(t *testing.T) {
	viper.Set(key.DownloaderAsync, false)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("page"))
	}))
	defer server.Close()

	chapter := &Chapter{Name: "stream", Index: 1}
	for i := 0; i < 3; i++ {
		chapter.Pages = append(chapter.Pages, &Page{
			Index:     uint16(i),
			URL:       server.URL,
			Extension: ".png",
			Chapter:   chapter,
		})
	}

	dir := "/tmp/chapters/stream"
	if err := chapter.DownloadPagesTo(dir, false, func(string) {}); err != nil {
		t.Fatalf("DownloadPagesTo failed: %v", err)
	}

	entries, err := filesystem.Api().ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 page files, got %d", len(entries))
	}

	for _, page := range chapter.Pages {
		if page.Contents != nil {
			t.Error("streamed pages must not buffer contents in memory")
		}
	}
}
