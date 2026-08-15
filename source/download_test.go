package source

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
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
