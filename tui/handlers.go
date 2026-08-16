package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yukiteruamano/koma/anilist"
	"github.com/yukiteruamano/koma/color"
	"github.com/yukiteruamano/koma/downloader"
	"github.com/yukiteruamano/koma/log"
	"github.com/yukiteruamano/koma/provider"
	"github.com/yukiteruamano/koma/source"
	"github.com/yukiteruamano/koma/style"
	"github.com/yukiteruamano/koma/util"
	"sync"
)

func (b *statefulBubble) loadSources(ps []*provider.Provider) tea.Cmd {
	return func() tea.Msg {
		sources := make([]source.Source, len(ps))
		wg := sync.WaitGroup{}
		wg.Add(len(ps))

		for i, p := range ps {
			go func(i int, p *provider.Provider) {
				defer wg.Done()

				log.Info("loading source " + p.ID)
				s, err := p.CreateSource()

				if err != nil {
					log.Error(err)
					b.errorChannel <- err
					return
				}

				log.Info("source " + p.ID + " loaded")
				sources[i] = s
			}(i, p)
		}

		wg.Wait()

		// drop sources that failed to load so the list has no nil entries
		valid := sources[:0]
		for _, s := range sources {
			if s != nil {
				valid = append(valid, s)
			}
		}

		b.sourcesLoadedChannel <- valid

		return nil
	}
}

func (b *statefulBubble) waitForSourcesLoaded() tea.Cmd {
	return func() tea.Msg {
		select {
		case res := <-b.sourcesLoadedChannel:
			return res
		case err := <-b.errorChannel:
			return err
		}
	}
}

func (b *statefulBubble) searchManga(query string) tea.Cmd {
	return func() tea.Msg {
		log.Info("searching for " + query)

		sources := b.selectedSources
		results := make([][]*source.Manga, len(sources))

		wg := sync.WaitGroup{}
		wg.Add(len(sources))
		for i, s := range sources {
			go func(i int, s source.Source) {
				defer wg.Done()
				sourceMangas, err := s.Search(query)

				if err != nil {
					log.Error(err)
					b.errorChannel <- err
					return
				}

				log.Infof("found %s from source %s", util.Quantify(len(sourceMangas), "manga", "mangas"), s.Name())
				results[i] = sourceMangas
			}(i, s)
		}

		wg.Wait()

		var mangas []*source.Manga
		for _, result := range results {
			// a provider may have failed and left its slot nil
			mangas = append(mangas, result...)
		}

		log.Infof("found %d mangas from %d sources", len(mangas), len(sources))

		b.foundMangasChannel <- mangas

		return nil
	}
}

func (b *statefulBubble) waitForMangas() tea.Cmd {
	return func() tea.Msg {
		select {
		case found := <-b.foundMangasChannel:
			return found
		case err := <-b.errorChannel:
			return err
		}
	}
}

func (b *statefulBubble) getChapters(manga *source.Manga) tea.Cmd {
	return func() tea.Msg {
		log.Info("getting chapters of " + manga.Name)
		chapters, err := manga.Source.ChaptersOf(manga)
		if err != nil {
			log.Error(err)
			b.errorChannel <- err
		} else {
			log.Infof("found %s", util.Quantify(len(chapters), "chapter", "chapters"))
			b.foundChaptersChannel <- chapters
		}

		return nil
	}
}

func (b *statefulBubble) waitForChapters() tea.Cmd {
	return func() tea.Msg {
		select {
		case found := <-b.foundChaptersChannel:
			return found
		case err := <-b.errorChannel:
			return err
		}
	}
}

func (b *statefulBubble) readChapter(chapter *source.Chapter) tea.Cmd {
	return func() tea.Msg {
		err := downloader.Read(chapter, func(s string) {
			select {
			case b.progressChannel <- progressMsg{status: s}:
			default:
			}
		})

		if err != nil {
			b.errorChannel <- err
		} else {
			b.chapterReadChannel <- struct{}{}
		}

		// release any pending waitForProgress command
		b.progressDoneOnce.Do(func() { close(b.progressDone) })

		return nil
	}
}

func (b *statefulBubble) waitForChapterRead() tea.Cmd {
	return func() tea.Msg {
		select {
		case res := <-b.chapterReadChannel:
			return res
		case err := <-b.errorChannel:
			return err
		}
	}
}

func (b *statefulBubble) downloadChapter(chapter *source.Chapter) tea.Cmd {
	return func() tea.Msg {
		_, err := downloader.Download(chapter, func(s string) {
			select {
			case b.progressChannel <- progressMsg{status: s}:
			default:
				// drop if the UI is not draining progress fast enough
			}
		})

		return chapterDownloadResult{chapter: chapter, err: err}
	}
}

func (b *statefulBubble) waitForProgress() tea.Cmd {
	return func() tea.Msg {
		select {
		case msg := <-b.progressChannel:
			return msg
		case <-b.progressDone:
			return progressDoneMsg{}
		}
	}
}

// resetProgressDone recreates the progress-done signal so a second download
// batch (e.g. retrying failed chapters) gets a fresh channel.
func (b *statefulBubble) resetProgressDone() {
	b.progressDone = make(chan struct{})
	b.progressDoneOnce = sync.Once{}
}

func (b *statefulBubble) fetchAndSetAnilist(manga *source.Manga) tea.Cmd {
	return func() tea.Msg {
		alManga, err := anilist.FindClosest(manga.Name)
		if err != nil {
			// this error is not that important, we can ignore t
			log.Warn(err)
			alManga = nil
		}

		// always send, so waitForAnilistFetchAndSet does not block forever
		b.closestAnilistMangaChannel <- alManga
		return nil
	}
}

func (b *statefulBubble) waitForAnilistFetchAndSet() tea.Cmd {
	return func() tea.Msg {
		return <-b.closestAnilistMangaChannel
	}
}

func (b *statefulBubble) fetchAnilist(manga *source.Manga) tea.Cmd {
	return func() tea.Msg {
		log.Info("fetching anilist for " + manga.Name)
		b.progressStatus = fmt.Sprintf("Fetching anilist for %s", style.Fg(color.Purple)(manga.Name))
		mangas, err := anilist.SearchByName(manga.Name)
		if err != nil {
			log.Error(err)
			b.errorChannel <- err
		} else {
			log.Infof("found %s", util.Quantify(len(mangas), "manga", "mangas"))
			b.fetchedAnilistMangasChannel <- mangas
		}

		return nil
	}
}

func (b *statefulBubble) waitForAnilist() tea.Cmd {
	return func() tea.Msg {
		select {
		case found := <-b.fetchedAnilistMangasChannel:
			return found
		case err := <-b.errorChannel:
			return err
		}
	}
}

func (b *statefulBubble) selectChapterBy(f func(chapter *source.Chapter) bool) tea.Cmd {
	return func() tea.Msg {
		for i, item := range b.chaptersC.Items() {
			chapter := item.(*listItem).internal.(*source.Chapter)
			if f(chapter) {
				b.chaptersC.Select(i)
				return nil
			}
		}

		return nil
	}
}
