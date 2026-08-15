package query

import (
	"cmp"
	"slices"
	"sync"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/samber/mo"
	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/key"
)

var (
	mu sync.RWMutex

	// suggestionCache memoizes the matched suggestions per query prefix.
	// Stored as copied strings so concurrent Remember calls can never alias
	// shared record pointers.
	suggestionCache = make(map[string][]string)

	// sortedRecords holds a snapshot of the ranked query history, loaded once
	// and invalidated whenever Remember writes new data.
	sortedRecords []*queryRecord
	recordsFresh  bool
)

// maxSuggestionCache bounds the in-memory suggestion cache for long TUI sessions.
const maxSuggestionCache = 100

// invalidateSuggestions clears the suggestion state. Called from Remember after
// the query history is mutated.
func invalidateSuggestions() {
	mu.Lock()
	suggestionCache = make(map[string][]string)
	sortedRecords = nil
	recordsFresh = false
	mu.Unlock()
}

// loadRecordsLocked snapshots the ranked query history. Caller must hold mu.
func loadRecordsLocked() {
	recordsFresh = true

	cached, expired, err := cacher.Get()
	if err != nil || expired || cached == nil {
		sortedRecords = nil
		return
	}

	records := make([]*queryRecord, 0, len(cached))
	for _, record := range cached {
		records = append(records, &queryRecord{Rank: record.Rank, Query: record.Query})
	}

	slices.SortFunc(records, func(a, b *queryRecord) int {
		return cmp.Compare(b.Rank, a.Rank)
	})

	sortedRecords = records
}

func SuggestMany(query string) []string {
	if !viper.GetBool(key.SearchShowQuerySuggestions) {
		return []string{}
	}

	query = sanitize(query)

	mu.RLock()
	if cached, ok := suggestionCache[query]; ok {
		mu.RUnlock()
		return cached
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	// re-check now that we hold the write lock
	if cached, ok := suggestionCache[query]; ok {
		return cached
	}

	if len(suggestionCache) > maxSuggestionCache {
		suggestionCache = make(map[string][]string)
	}

	if !recordsFresh {
		loadRecordsLocked()
	}

	var matches []string
	for _, record := range sortedRecords {
		if fuzzy.Match(query, record.Query) {
			matches = append(matches, record.Query)
		}
	}

	suggestionCache[query] = matches
	return matches
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
