package source

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yukiteruamano/koma/filesystem"
)

func TestPageDownloadWithoutURL(t *testing.T) {
	tests := []struct {
		name     string
		contents []byte
		size     uint64
		wantErr  bool
	}{
		{
			name:     "contents present skips download",
			contents: []byte("data"),
			size:     4,
			wantErr:  false,
		},
		{
			name:    "no contents returns error",
			wantErr: true,
		},
		{
			name:     "zero size with contents returns error",
			contents: []byte("data"),
			size:     0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{
				Index:    1,
				Size:     tt.size,
				Contents: nil,
				Chapter:  &Chapter{},
			}

			if tt.contents != nil {
				page.Contents = bytes.NewBuffer(tt.contents)
			}

			err := page.Download()
			if tt.wantErr && err == nil {
				t.Fatal("expected an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestPageDownloadHTTPErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	page := &Page{
		URL:     server.URL,
		Index:   1,
		Chapter: &Chapter{URL: server.URL},
	}

	if err := page.Download(); err == nil {
		t.Fatal("expected an HTTP error for 404")
	}
}

func TestPageDownloadSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chunked" {
			w.Header().Set("Content-Type", "image/png")
			flusher, _ := w.(http.Flusher)
			_, _ = w.Write([]byte("chunk1"))
			flusher.Flush()
			_, _ = w.Write([]byte("chunk2"))
			return
		}
		w.Header().Set("Content-Length", "4")
		_, _ = w.Write([]byte("data"))
	}))
	defer server.Close()

	tests := []struct {
		name     string
		url      string
		wantSize uint64
		wantBody string
	}{
		{name: "known content length", url: server.URL + "/page", wantSize: 4, wantBody: "data"},
		{name: "chunked unknown length", url: server.URL + "/chunked", wantSize: 12, wantBody: "chunk1chunk2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{URL: tt.url, Index: 1, Chapter: &Chapter{URL: server.URL}}

			if err := page.Download(); err != nil {
				t.Fatalf("Download failed: %v", err)
			}

			if page.Size != tt.wantSize {
				t.Errorf("Size = %d, want %d", page.Size, tt.wantSize)
			}
			if page.Contents.String() != tt.wantBody {
				t.Errorf("Contents = %q, want %q", page.Contents.String(), tt.wantBody)
			}
		})
	}
}

func TestPageDownloadEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "0")
	}))
	defer server.Close()

	page := &Page{URL: server.URL, Index: 1, Chapter: &Chapter{URL: server.URL}}
	if err := page.Download(); err == nil {
		t.Fatal("expected an error for an empty body")
	}
}

func TestPageDownloadRequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Referer") != "https://example.com/chapter" {
			t.Errorf("Referer = %q, want chapter URL", r.Header.Get("Referer"))
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("User-Agent header missing")
		}
		_, _ = w.Write([]byte("x"))
	}))
	defer server.Close()

	page := &Page{URL: server.URL, Index: 1, Chapter: &Chapter{URL: "https://example.com/chapter"}}
	if err := page.Download(); err != nil {
		t.Fatalf("Download failed: %v", err)
	}
}

func TestPageDownloadTo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("page-bytes"))
	}))
	defer server.Close()

	page := &Page{URL: server.URL, Index: 1, Chapter: &Chapter{URL: server.URL}}

	dst := "/tmp/pages/0000000001.png"
	if err := page.DownloadTo(dst); err != nil {
		t.Fatalf("DownloadTo failed: %v", err)
	}

	if page.Contents != nil {
		t.Error("DownloadTo must not buffer contents in memory")
	}
	if page.Size != 10 {
		t.Errorf("Size = %d, want 10", page.Size)
	}

	data, err := filesystem.Api().ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "page-bytes" {
		t.Errorf("file contents = %q, want page-bytes", data)
	}
}

func TestPageDownloadToHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	page := &Page{URL: server.URL, Index: 1, Chapter: &Chapter{URL: server.URL}}

	if err := page.DownloadTo("/tmp/pages/x.png"); err == nil {
		t.Fatal("expected an error for a 500 response")
	}
}

func TestPageFilenamePadding(t *testing.T) {
	tests := []struct {
		name      string
		index     uint16
		extension string
		want      string
	}{
		{name: "single digit", index: 1, extension: ".png", want: "000001.png"},
		{name: "large index", index: 999, extension: ".jpg", want: "000999.jpg"},
		{name: "no extension", index: 5, extension: "", want: "0000000005"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{Index: tt.index, Extension: tt.extension}
			if got := page.Filename(); got != tt.want {
				t.Errorf("Filename() = %q, want %q", got, tt.want)
			}
		})
	}
}
