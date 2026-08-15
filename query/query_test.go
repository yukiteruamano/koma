package query

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	filesystem.SetMemMapFs()
	viper.Set(key.SearchShowQuerySuggestions, true)
	goleak.VerifyTestMain(m,
		goleak.IgnoreAnyFunction("net/http.(*http2clientConnReadLoop).run"),
	)
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "trims whitespace", input: "  death note  ", want: "death note"},
		{name: "lowercases", input: "Death Note", want: "death note"},
		{name: "empty stays empty", input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitize(tt.input); got != tt.want {
				t.Errorf("sanitize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRememberIncrementsRank(t *testing.T) {
	const query = "one piece"

	// the disk cache may persist across -count runs, so assert the delta
	cached, _, err := cacher.Get()
	if err != nil {
		t.Fatal(err)
	}
	before := 0
	if record, ok := cached[query]; ok {
		before = record.Rank
	}

	if err := Remember(query, 1); err != nil {
		t.Fatalf("Remember failed: %v", err)
	}
	if err := Remember(query, 2); err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	cached, _, err = cacher.Get()
	if err != nil {
		t.Fatal(err)
	}

	record, ok := cached[query]
	if !ok {
		t.Fatal("expected query to be recorded")
	}

	if record.Rank != before+3 {
		t.Errorf("rank = %d, want %d (delta 3)", record.Rank, before+3)
	}
}

func TestRememberEvictsWhenOverCap(t *testing.T) {
	for i := 0; i < maxRemembered+50; i++ {
		if err := Remember(fmt.Sprintf("query-%d", i), 1); err != nil {
			t.Fatalf("Remember failed: %v", err)
		}
	}

	cached, _, err := cacher.Get()
	if err != nil {
		t.Fatal(err)
	}

	if len(cached) > maxRemembered {
		t.Errorf("cache has %d entries, want at most %d", len(cached), maxRemembered)
	}
}

func TestSuggestMatchesFuzzy(t *testing.T) {
	_ = Remember("death note", 10)
	_ = Remember("death parade", 5)

	suggestions := SuggestMany("death")
	if len(suggestions) == 0 {
		t.Fatal("expected suggestions for 'death'")
	}

	if suggestions[0] != "death note" {
		t.Errorf("expected highest-rank suggestion first, got %q", suggestions[0])
	}
}

func TestSuggestCached(t *testing.T) {
	_ = Remember("chainsaw man", 1)

	first := SuggestMany("chainsaw")
	if len(first) == 0 {
		t.Fatal("expected suggestions")
	}

	if _, ok := suggestionCache["chainsaw"]; !ok {
		t.Error("expected the prefix to be cached after the first call")
	}

	second := SuggestMany("chainsaw")
	if len(second) != len(first) {
		t.Errorf("cached suggestions differ: %v vs %v", first, second)
	}
}

func TestSuggestDisabledByConfig(t *testing.T) {
	viper.Set(key.SearchShowQuerySuggestions, false)
	defer viper.Set(key.SearchShowQuerySuggestions, true)

	if got := SuggestMany("anything"); len(got) != 0 {
		t.Errorf("expected no suggestions when disabled, got %v", got)
	}
}

func TestSuggestOption(t *testing.T) {
	_ = Remember("vinland saga", 5)

	if got := Suggest("vinland"); got.IsAbsent() {
		t.Fatal("expected a suggestion")
	}

	if got := Suggest("zzzz-nothing"); !got.IsAbsent() {
		t.Errorf("expected no suggestion for a non-matching query")
	}
}
