package generic

import (
	"github.com/gocolly/colly/v2"
	"github.com/yukiteruamano/koma/source"
	"net/http"
)

// ChaptersOf given source.Manga
func (s *Scraper) ChaptersOf(manga *source.Manga) ([]*source.Chapter, error) {
	s.mu.Lock()
	if chapters, ok := s.chapters[manga.URL]; ok {
		// rehydrate the Manga back-reference for a different Manga instance
		for _, chapter := range chapters {
			chapter.Manga = manga
		}
		manga.Chapters = chapters
		s.mu.Unlock()
		return chapters, nil
	}
	s.mu.Unlock()

	ctx := colly.NewContext()
	ctx.Put("manga", manga)
	err := s.chaptersCollector.Request(http.MethodGet, manga.URL, nil, ctx, nil)

	if err != nil {
		return nil, err
	}

	s.chaptersCollector.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config.ReverseChapters {
		// reverse chapters; indices stay 0-based like every other provider
		chapters := s.chapters[manga.URL]
		reversed := make([]*source.Chapter, len(chapters))
		for i, chapter := range chapters {
			reversed[len(chapters)-i-1] = chapter
			chapter.Index = uint16(len(chapters) - i - 1)
		}

		s.chapters[manga.URL] = reversed
		manga.Chapters = reversed
	}

	return s.chapters[manga.URL], nil
}
