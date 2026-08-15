//go:build integration

// Package anilist holds integration tests that hit the live Anilist GraphQL
// API. Run with: go test -tags=integration ./test/integration/...
package anilist

import (
	"testing"

	"github.com/yukiteruamano/koma/anilist"
)

func TestSearchByNameIntegration(t *testing.T) {
	results, err := anilist.SearchByName("Death Note")
	if err != nil {
		t.Fatalf("SearchByName failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	if results[0].Title.English != "Death Note" {
		t.Fatalf("expected first result to be Death Note, got %q", results[0].Title.English)
	}
}

func TestFindClosestIntegration(t *testing.T) {
	result, err := anilist.FindClosest("Death Note")
	if err != nil {
		t.Fatalf("FindClosest failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected a result")
	}

	if result.Title.English != "Death Note" {
		t.Fatalf("expected Death Note, got %q", result.Title.English)
	}
}
