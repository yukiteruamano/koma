package inline

import (
	"encoding/json"
	"io"

	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/anilist"
	"github.com/yukiteruamano/koma/key"
	"github.com/yukiteruamano/koma/source"
)

type Manga struct {
	// Source that the manga belongs to.
	Source string `json:"source" jsonschema:"description=Source that the manga belongs to."`
	// Koma variant of the manga
	Koma *source.Manga `json:"koma" jsonschema:"description=Koma variant of the manga"`
	// Anilist is the closest anilist match to koma manga
	Anilist *anilist.Manga `json:"anilist" jsonschema:"description=Anilist is the closest anilist match to koma manga"`
}

type Output struct {
	Query  string   `json:"query" jsonschema:"description=Query that was used to search for the manga."`
	Result []*Manga `json:"result" jsonschema:"description=Result of the search."`
}

func asJson(manga []*source.Manga, options *Options, w io.Writer) error {
	var m = make([]*Manga, len(manga))
	for i, manga := range manga {
		al := manga.Anilist.OrElse(nil)
		if !options.IncludeAnilistManga {
			al = nil
		}

		m[i] = &Manga{
			Koma:    manga,
			Anilist: al,
			Source:  manga.Source.Name(),
		}
	}

	// Stream the output instead of buffering the whole document in memory.
	return json.NewEncoder(w).Encode(&Output{
		Result: m,
		Query:  options.Query,
	})
}

func prepareManga(manga *source.Manga, options *Options) error {
	var err error

	if options.IncludeAnilistManga {
		err = manga.BindWithAnilist()
		if err != nil {
			return err
		}
	}

	if options.ChaptersFilter.IsPresent() {
		chapters, err := manga.Source.ChaptersOf(manga)
		if err != nil {
			return err
		}

		chapters, err = options.ChaptersFilter.MustGet()(chapters)
		if err != nil {
			return err
		}

		manga.Chapters = chapters

		if options.PopulatePages {
			for _, chapter := range chapters {
				_, err := chapter.Source().PagesOf(chapter)
				if err != nil {
					return err
				}
			}
		}
	} else {
		// clear chapters in case they were loaded from cache or something
		manga.Chapters = make([]*source.Chapter, 0)
	}

	if viper.GetBool(key.MetadataFetchAnilist) {
		_ = manga.PopulateMetadata(func(string) {})
	}

	return nil
}
