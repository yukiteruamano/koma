package zonatmo

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/yukiteruamano/koma/constant"
	"github.com/yukiteruamano/koma/network"
	"github.com/yukiteruamano/koma/source"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

const (
	ID   = "zonatmo-built-in"
	Name = "Zonatmo"
)

// baseURL is a var so tests can point the source at a local server.
var baseURL = "https://zonatmo.org"

type Source struct {
	cache struct {
		mangas   *cacher[[]*source.Manga]
		chapters *cacher[[]*source.Chapter]
	}
}

func New() *Source {
	s := &Source{}
	s.cache.mangas = newCacher[[]*source.Manga]("zonatmo_mangas")
	s.cache.chapters = newCacher[[]*source.Chapter]("zonatmo_chapters")
	return s
}

func (s *Source) Name() string { return Name }
func (s *Source) ID() string   { return ID }

// absoluteURL resolves a possibly-relative href against the site base URL.
func absoluteURL(href string) string {
	if href == "" {
		return href
	}
	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}
	if u, err := url.Parse(href); err == nil && u.IsAbs() {
		return href
	}
	return baseURL + href
}

// urlID extracts the trailing path segment of a (possibly relative) URL.
func urlID(href string) string {
	u, err := url.Parse(absoluteURL(href))
	if err != nil {
		return ""
	}
	return path.Base(u.Path)
}

func (s *Source) newRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", constant.RandomUserAgent())
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9")
	return req, nil
}

func (s *Source) fetchDocument(req *http.Request) (*goquery.Document, error) {
	resp, err := network.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zonatmo: %s returned %d", req.URL, resp.StatusCode)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

func (s *Source) Search(query string) ([]*source.Manga, error) {
	if cached, ok := s.cache.mangas.Get(query).Get(); ok {
		// Manga.Source is json:"-", so rehydrate it after loading from the cache
		for _, manga := range cached {
			manga.Source = s
		}
		return cached, nil
	}

	searchURL := fmt.Sprintf("%s/biblioteca?title=%s&_pg=1", baseURL, url.QueryEscape(query))

	req, err := s.newRequest("GET", searchURL)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", baseURL)

	doc, err := s.fetchDocument(req)
	if err != nil {
		return nil, err
	}

	var mangas []*source.Manga
	doc.Find("a[href*='/library/']").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists || href == "" {
			return
		}

		name := strings.TrimSpace(sel.Find("h4[title]").First().AttrOr("title", ""))
		if name == "" {
			name = strings.TrimSpace(sel.Find("h4").First().Text())
		}
		if name == "" {
			return
		}

		cover := sel.Find("img[src]").First().AttrOr("src", "")

		manga := &source.Manga{
			Name:   name,
			URL:    absoluteURL(href),
			Index:  uint16(i),
			ID:     urlID(href),
			Source: s,
		}
		manga.Metadata.Cover.ExtraLarge = absoluteURL(cover)

		mangas = append(mangas, manga)
	})

	_ = s.cache.mangas.Set(query, mangas)
	return mangas, nil
}

func (s *Source) ChaptersOf(manga *source.Manga) ([]*source.Chapter, error) {
	if cached, ok := s.cache.chapters.Get(manga.ID).Get(); ok {
		// Chapter.Manga is json:"-", so rehydrate it after loading from the cache
		for _, chapter := range cached {
			chapter.Manga = manga
		}
		manga.Chapters = cached
		return cached, nil
	}

	req, err := s.newRequest("GET", manga.URL)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", baseURL)

	doc, err := s.fetchDocument(req)
	if err != nil {
		return nil, err
	}

	var chapters []*source.Chapter
	doc.Find("li.upload-link").Each(func(i int, sel *goquery.Selection) {
		name := strings.TrimSpace(sel.Find(".chapter-number").First().Text())
		if name == "" {
			return
		}

		readerHref, exists := sel.Find(".chapter-detail a[href*='view_uploads/']").First().Attr("href")
		if !exists || readerHref == "" {
			return
		}

		chapters = append(chapters, &source.Chapter{
			Name:  name,
			URL:   absoluteURL(readerHref),
			ID:    urlID(readerHref),
			Manga: manga,
		})
	})

	// Chapters are listed newest first; reverse to ascending order.
	for i, j := 0, len(chapters)-1; i < j; i, j = i+1, j-1 {
		chapters[i], chapters[j] = chapters[j], chapters[i]
	}

	for i, ch := range chapters {
		ch.Index = uint16(i)
	}

	manga.Chapters = chapters
	_ = s.cache.chapters.Set(manga.ID, chapters)
	return chapters, nil
}

func (s *Source) PagesOf(chapter *source.Chapter) ([]*source.Page, error) {
	req, err := s.newRequest("GET", chapter.URL)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", baseURL)

	doc, err := s.fetchDocument(req)
	if err != nil {
		return nil, err
	}

	var pages []*source.Page
	doc.Find("img.reader-image").Each(func(i int, sel *goquery.Selection) {
		src, exists := sel.Attr("src")
		if !exists || src == "" {
			return
		}

		ext := filepath.Ext(src)
		ext = strings.Split(ext, "?")[0]

		pages = append(pages, &source.Page{
			URL:       absoluteURL(src),
			Index:     uint16(i),
			Extension: ext,
			Chapter:   chapter,
		})
	})

	chapter.Pages = pages
	return pages, nil
}
