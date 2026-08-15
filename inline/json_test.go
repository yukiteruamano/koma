package inline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/yukiteruamano/koma/source"
	"github.com/yukiteruamano/koma/test/testutil"
)

func TestAsJsonShape(t *testing.T) {
	fake := &testutil.FakeSource{NameValue: "mocksource"}
	manga := &source.Manga{Name: "Mock Manga", Source: fake}

	var out bytes.Buffer
	options := &Options{
		Out:   &out,
		Query: "mock",
	}

	err := asJson([]*source.Manga{manga}, options, &out)
	if err != nil {
		t.Fatalf("asJson failed: %v", err)
	}

	var decoded struct {
		Query  string `json:"query"`
		Result []struct {
			Source string      `json:"source"`
			Koma   interface{} `json:"koma"`
		} `json:"result"`
	}

	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if decoded.Query != "mock" {
		t.Errorf("query = %q, want mock", decoded.Query)
	}

	if len(decoded.Result) != 1 {
		t.Fatalf("result has %d entries, want 1", len(decoded.Result))
	}

	if decoded.Result[0].Source != "mocksource" {
		t.Errorf("source = %q, want mocksource", decoded.Result[0].Source)
	}

	if decoded.Result[0].Koma == nil {
		t.Error("expected koma variant to be present")
	}
}

func TestAsJsonEmpty(t *testing.T) {
	var out bytes.Buffer
	options := &Options{Out: &out, Query: "empty"}

	if err := asJson(nil, options, &out); err != nil {
		t.Fatalf("asJson failed: %v", err)
	}

	var decoded struct {
		Result []interface{} `json:"result"`
	}

	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if decoded.Result == nil {
		t.Error("expected result to be an empty array, not null")
	}
}
