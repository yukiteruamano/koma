package generic

import (
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
	return s.mangas[address], nil
}
