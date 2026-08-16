package mini

import (
	"errors"
	"github.com/samber/lo"
	"github.com/yukiteruamano/koma/source"
	"github.com/yukiteruamano/koma/util"
	"os"
)

var (
	truncateAt = 100
)

type Options struct {
	Download bool
	Continue bool
}

type mini struct {
	state         state
	statesHistory util.Stack[state]

	download bool

	selectedSource source.Source

	cachedMangas   map[string][]*source.Manga
	cachedChapters map[string][]*source.Chapter

	query            string
	selectedManga    *source.Manga
	selectedChapters []*source.Chapter
}

func newMini() *mini {
	return &mini{
		statesHistory:  util.Stack[state]{},
		cachedMangas:   make(map[string][]*source.Manga),
		cachedChapters: make(map[string][]*source.Chapter),
	}
}

func (m *mini) previousState() {
	if m.statesHistory.Len() > 0 {
		m.setState(m.statesHistory.Pop())
	}
}

func (m *mini) setState(s state) {
	m.state = s
}

func (m *mini) newState(s state) {
	// do not push state if it is the same as the current state
	if m.state == s {
		return
	}

	// Transitioning to these states is not allowed (it makes no sense)
	if !lo.Contains([]state{}, m.state) {
		m.statesHistory.Push(m.state)
	}

	m.setState(s)
}

func Run(options *Options) error {
	if options.Continue && options.Download {
		return errors.New("cannot download and continue")
	}

	m := newMini()
	m.state = sourceSelectState
	if options.Continue {
		m.state = historySelectState
	}

	m.download = options.Download

	if w, _, err := util.TerminalSize(); err == nil {
		truncateAt = w
	}

	var err error

	for {
		if err = m.handleState(); err != nil {
			return err
		}
	}
}

func (m *mini) handleState() error {
	switch m.state {
	case historySelectState:
		return m.handleHistorySelectState()
	case sourceSelectState:
		return m.handleSourceSelectState()
	case mangasSearchState:
		return m.handleMangaSearchState()
	case mangaSelectState:
		return m.handleMangaSelectState()
	case chapterSelectState:
		return m.handleChapterSelectState()
	case chapterReadState:
		return m.handleChapterReadState()
	case chaptersDownloadState:
		return m.handleChaptersDownloadState()
	case quitState:
		if m.selectedSource != nil {
			source.CloseSource(m.selectedSource)
		}
		os.Exit(0)
	}

	return nil
}
