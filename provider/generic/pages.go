package generic

import (
	"github.com/gocolly/colly/v2"
	"github.com/yukiteruamano/koma/source"
	"net/http"
)

// PagesOf given source.Chapter
func (s *Scraper) PagesOf(chapter *source.Chapter) ([]*source.Page, error) {
	s.mu.Lock()
	if pages, ok := s.pages[chapter.URL]; ok {
		s.mu.Unlock()
		return pages, nil
	}
	s.mu.Unlock()

	ctx := colly.NewContext()
	ctx.Put("chapter", chapter)
	err := s.pagesCollector.Request(http.MethodGet, chapter.URL, nil, ctx, nil)

	if err != nil {
		return nil, err
	}

	s.pagesCollector.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pages[chapter.URL], nil
}
