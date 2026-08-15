package mangadex

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/darylhjd/mangodex"
	"github.com/yukiteruamano/koma/source"
)

func (m *Mangadex) PagesOf(chapter *source.Chapter) ([]*source.Page, error) {
	u, _ := url.Parse(mangodex.BaseAPI)
	u.Path = fmt.Sprintf(mangodex.GetMDHomeURLPath, chapter.ID)

	var server mangodex.MDHomeServerResponse
	if err := m.client.RequestAndDecode(context.Background(), http.MethodGet, u.String(), nil, &server); err != nil {
		return nil, err
	}

	names := server.Chapter.Data
	if len(names) == 0 {
		return nil, errors.New("there were no pages for this chapter")
	}

	// Build page URLs only. Contents are fetched lazily by the downloader,
	// so pages download concurrently through the generic async machinery.
	var pages = make([]*source.Page, len(names))
	for i, name := range names {
		pages[i] = &source.Page{
			Index:     uint16(i),
			URL:       strings.Join([]string{server.BaseURL, "data", server.Chapter.Hash, name}, "/"),
			Chapter:   chapter,
			Extension: filepath.Ext(name),
		}
	}

	chapter.Pages = pages
	return pages, nil
}
