package version

import "testing"

func TestCompareVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "patch lower", a: "1.0.0", b: "1.0.1", want: -1},
		{name: "patch higher", a: "1.0.1", b: "1.0.0", want: 1},
		{name: "minor lower", a: "1.0.0", b: "1.1.0", want: -1},
		{name: "minor higher", a: "1.1.0", b: "1.0.0", want: 1},
		{name: "major lower", a: "1.0.0", b: "2.0.0", want: -1},
		{name: "major higher", a: "2.0.0", b: "1.0.0", want: 1},
		{name: "equal", a: "1.0.0", b: "1.0.0", want: 0},
		{name: "v prefix", a: "v1.2.3", b: "1.2.3", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Compare(tt.a, tt.b)
			if err != nil {
				t.Fatalf("Compare(%q, %q) failed: %v", tt.a, tt.b, err)
			}
			if got != tt.want {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCompareVersionsInvalid(t *testing.T) {
	t.Parallel()

	if _, err := Compare("not-a-version", "1.0.0"); err == nil {
		t.Error("expected an error for an invalid version")
	}
}
