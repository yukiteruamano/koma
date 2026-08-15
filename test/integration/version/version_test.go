//go:build integration

// Package version holds integration tests that query the live GitHub API.
// Run with: go test -tags=integration ./test/integration/...
package version

import (
	"regexp"
	"testing"

	"github.com/yukiteruamano/koma/version"
)

func TestLatestVersionIntegration(t *testing.T) {
	latest, err := version.Latest()
	if err != nil {
		t.Fatalf("Latest failed: %v", err)
	}

	if latest == "" {
		t.Fatal("expected a non-empty version")
	}

	semver := regexp.MustCompile(`^v?(\d+)(\.\d+){0,2}(-\w+)?$`)
	if !semver.MatchString(latest) {
		t.Fatalf("expected semver, got %q", latest)
	}
}
