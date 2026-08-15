// Package testutil provides shared helpers for koma tests: tiny PNG fixtures
// and a fake source.Source backed by an in-memory HTTP server.
package testutil

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yukiteruamano/koma/source"
)

// PNGImage returns a valid PNG image of the given dimensions.
func PNGImage(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// NewImageServer starts an HTTP server that serves a PNG image (or the given
// status code) for any request. The server is closed when the test finishes.
func NewImageServer(t *testing.T, status int) *httptest.Server {
	t.Helper()

	img := PNGImage(100, 150)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(img)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// PagesFromServer builds pageCount source.Pages whose URLs point at the server.
func PagesFromServer(server *httptest.Server, chapter *source.Chapter, pageCount int) []*source.Page {
	pages := make([]*source.Page, pageCount)
	for i := 0; i < pageCount; i++ {
		pages[i] = &source.Page{
			URL:       server.URL + "/page.png",
			Index:     uint16(i),
			Chapter:   chapter,
			Extension: ".png",
		}
	}

	return pages
}

// FakeSource is a controllable source.Source implementation.
type FakeSource struct {
	NameValue string
	SearchOut []*source.Manga
	Chapters  []*source.Chapter
	Pages     []*source.Page
	SearchErr error
}

func (f *FakeSource) Name() string {
	if f.NameValue == "" {
		return "fake"
	}
	return f.NameValue
}

func (f *FakeSource) ID() string {
	return "fake"
}

func (f *FakeSource) Search(string) ([]*source.Manga, error) {
	return f.SearchOut, f.SearchErr
}

func (f *FakeSource) ChaptersOf(*source.Manga) ([]*source.Chapter, error) {
	return f.Chapters, nil
}

func (f *FakeSource) PagesOf(*source.Chapter) ([]*source.Page, error) {
	return f.Pages, nil
}

// NewChapter builds a chapter wired to a fake source with pages served by an
// in-memory image server.
func NewChapter(t *testing.T, name string, pageCount int) (*source.Chapter, *FakeSource, *httptest.Server) {
	t.Helper()

	server := NewImageServer(t, http.StatusOK)

	fake := &FakeSource{NameValue: "fake"}
	manga := &source.Manga{Name: name, Source: fake}
	chapter := &source.Chapter{Name: name, Index: 1, Manga: manga}
	chapter.Pages = PagesFromServer(server, chapter, pageCount)
	fake.Pages = chapter.Pages
	fake.Chapters = []*source.Chapter{chapter}

	return chapter, fake, server
}

// ChapterWithPages builds a chapter whose pages already hold in-memory PNG
// contents, for converter tests that do not need the network.
func ChapterWithPages(name string, pageCount int) *source.Chapter {
	manga := &source.Manga{Name: name}
	chapter := &source.Chapter{Name: name, Index: 1, Manga: manga}

	img := PNGImage(100, 150)
	for i := 0; i < pageCount; i++ {
		chapter.Pages = append(chapter.Pages, &source.Page{
			Index:     uint16(i),
			Extension: ".png",
			Chapter:   chapter,
			Contents:  bytes.NewBuffer(img),
		})
	}

	return chapter
}
