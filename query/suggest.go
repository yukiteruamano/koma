package query

import (
	"cmp"
	"slices"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/key"
)

var (
	suggestionCache = make(map[string][]*queryRecord)
)

// maxSuggestionCache bounds the in-memory suggestion cache for long TUI sessions.
const maxSuggestionCache = 100

func SuggestMany(query string) []string {
	if !viper.GetBool(key.SearchShowQuerySuggestions) {
		return []string{}
	}

	query = sanitize(query)

	// keep the per-session cache bounded
	if len(suggestionCache) > maxSuggestionCache {
		suggestionCache = make(map[string][]*queryRecord)
	}

	var records []*queryRecord

	if prev, ok := suggestionCache[query]; ok {
		records = prev
	} else {
		cached, expired, err := cacher.Get()
		if err != nil || expired || cached == nil {
			return []string{}
		}

		for _, record := range cached {
			if fuzzy.Match(query, record.Query) {
				records = append(records, record)
			}
		}

		slices.SortFunc(records, func(a, b *queryRecord) int {
			return cmp.Compare(b.Rank, a.Rank)
		})

		suggestionCache[query] = records
	}

	return lo.Map(records, func(record *queryRecord, _ int) string {
		return record.Query
	})
}

// Suggest gives a suggestion for a query
func Suggest(query string) mo.Option[string] {
	records := SuggestMany(query)

	var suggestion mo.Option[string]

	if len(records) == 0 {
		suggestion = mo.None[string]()
	} else {
		suggestion = mo.Some(records[0])
	}

	return suggestion
}
