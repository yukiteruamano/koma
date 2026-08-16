package tui

import (
	"fmt"
	"github.com/yukiteruamano/koma/anilist"
	"github.com/yukiteruamano/koma/history"
	"github.com/yukiteruamano/koma/icon"
	"github.com/yukiteruamano/koma/provider"
	"github.com/yukiteruamano/koma/source"
	"github.com/yukiteruamano/koma/style"
	"strings"
)

type listItem struct {
	internal interface{}
	marked   bool
}

func (t *listItem) toggleMark() {
	t.marked = !t.marked
}

func (t *listItem) getMark() string {
	switch t.internal.(type) {
	case *source.Chapter:
		return style.Bold(icon.Get(icon.Mark))
	case *anilist.Manga:
		return icon.Get(icon.Link)
	case *provider.Provider:
		return icon.Get(icon.Search)
	default:
		return ""
	}
}

func (t *listItem) Title() (title string) {
	switch e := t.internal.(type) {
	case *source.Chapter:
		var sb = strings.Builder{}

		sb.WriteString(t.FilterValue())
		if e.Volume != "" {
			sb.WriteString(" ")
			sb.WriteString(style.Faint(e.Volume))
		}

		if e.IsDownloaded() {
			sb.WriteString(" ")
			sb.WriteString(icon.Get(icon.Downloaded))
		}

		title = sb.String()
	default:
		title = t.FilterValue()
	}

	if title != "" && t.marked {
		//title = fmt.Sprintf("%s %s", title, icon.Get(icon.Mark))
		title = fmt.Sprintf("%s %s", title, t.getMark())
	}

	return
}

func (t *listItem) Description() (description string) {
	switch e := t.internal.(type) {
	case *source.Chapter:
		description = e.URL
	case *source.Manga:
		description = e.URL
	case *history.SavedChapter:
		description = fmt.Sprintf("%s : %d / %d", e.Name, e.Index, e.MangaChaptersTotal)
	case *provider.Provider:
		description = "Builtin"
	case *anilist.Manga:
		description = e.SiteURL
	}

	return
}

func (t *listItem) FilterValue() string {
	switch e := t.internal.(type) {
	case *source.Chapter:
		return e.Name
	case *source.Manga:
		return e.Name
	case *history.SavedChapter:
		return e.MangaName
	case *anilist.Manga:
		return e.Name()
	case *provider.Provider:
		return e.Name
	default:
		return ""
	}
}
