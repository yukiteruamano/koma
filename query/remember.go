package query

import (
	"sort"
	"sync"
)

// rememberMu serializes the Get-mutate-Set cycle so concurrent calls
// (TUI goroutines + metadata fetching) do not lose rank updates.
var rememberMu sync.Mutex

// maxRemembered bounds the size of queries.json so it does not grow forever.
const maxRemembered = 200

// Remember will add a query to the history.
// If query is already in the history, it will increment the rank by given weight
func Remember(query string, weight int) error {
	query = sanitize(query)

	rememberMu.Lock()
	defer rememberMu.Unlock()

	cached, expired, err := cacher.Get()
	if expired || err != nil {
		cached = map[string]*queryRecord{}
	}

	if cached == nil {
		cached = make(map[string]*queryRecord)
	}

	// if the query is already in the cache
	// increment its rank
	if record, ok := cached[query]; ok {
		record.Rank += weight
	} else {
		cached[query] = &queryRecord{
			Rank:  weight,
			Query: query,
		}
	}

	// evict the lowest-ranked queries when the cache grows too large
	if len(cached) > maxRemembered {
		records := make([]*queryRecord, 0, len(cached))
		for _, record := range cached {
			records = append(records, record)
		}
		sort.Slice(records, func(i, j int) bool { return records[i].Rank < records[j].Rank })

		for _, record := range records[:len(records)-maxRemembered] {
			delete(cached, record.Query)
		}
	}

	if err := cacher.Set(cached); err != nil {
		return err
	}

	invalidateSuggestions()
	return nil
}
