package generic

import (
	"fmt"

	"github.com/yukiteruamano/koma/source"
)

// Search for mangas by given title
func (s *Scraper) Search(query string) ([]*source.Manga, error) {
	address := s.config.GenerateSearchURL(query)

	s.mu.Lock()
	urls, ok := s.mangas[address]
	s.mu.Unlock()
	if ok {
		return urls, nil
	}

	err := s.mangasCollector.Visit(address)

	if err != nil {
		return nil, err
	}

	s.mangasCollector.Wait()
	s.mu.Lock()
	defer s.mu.Unlock()

	urls, ok = s.mangas[address]
	if !ok {
		// the request may have been redirected and stored under a different key
		return nil, fmt.Errorf("no search results for %q", query)
	}

	return urls, nil
}
