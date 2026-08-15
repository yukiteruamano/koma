package util

import (
	. "github.com/smartystreets/goconvey/convey"
	"regexp"
	"testing"
)

func TestPadZero(t *testing.T) {
	num := "123"
	Convey("Given a string "+num, t, func() {
		Convey("When padding with 4 zeros", func() {
			result := PadZero(num, 4)
			Convey("Then the result should be 0123", func() {
				So(result, ShouldEqual, "0123")
			})
		})

		Convey("When padding with 3 zeros", func() {
			result := PadZero(num, 3)
			Convey("Then the result should be 123", func() {
				So(result, ShouldEqual, "123")
			})
		})

		Convey("When padding with 2 zeros", func() {
			result := PadZero(num, 2)
			Convey("Then the result should be 123", func() {
				So(result, ShouldEqual, "123")
			})
		})

		Convey("When negative padding is performed", func() {
			result := PadZero(num, -1)
			Convey("Then the result should be 123", func() {
				So(result, ShouldEqual, "123")
			})
		})
	})
}

func TestFileStem(t *testing.T) {
	Convey("When the file name is 'foo.bar'", t, func() {
		result := FileStem("foo.bar")
		Convey("Then the result should be 'foo'", func() {
			So(result, ShouldEqual, "foo")
		})
	})
	Convey("When the file name is 'foo'", t, func() {
		result := FileStem("foo")
		Convey("Then the result should be 'foo'", func() {
			So(result, ShouldEqual, "foo")
		})
	})
	Convey("When the file name is 'foo.bar.baz'", t, func() {
		result := FileStem("foo.bar.baz")
		Convey("Then the result should be 'foo.bar'", func() {
			So(result, ShouldEqual, "foo.bar")
		})
	})
}

func TestQuantity(t *testing.T) {
	var (
		singular = "singular"
		plural   = "plural"
	)

	Convey("Given a quantity of 1", t, func() {
		quantity := 1
		Convey("When the quantity is converted to a string", func() {
			result := Quantify(quantity, singular, plural)
			Convey("Then the result should be '1 singular'", func() {
				So(result, ShouldEqual, "1 "+singular)
			})
		})
	})

	Convey("Given a quantity of 2", t, func() {
		quantity := 2
		Convey("When the quantity is converted to a string", func() {
			result := Quantify(quantity, singular, plural)
			Convey("Then the result should be '2 plural'", func() {
				So(result, ShouldEqual, "2 "+plural)
			})
		})
	})
}

func TestSanitizeFilename(t *testing.T) {
	invalidFilename := "~C:invalid/file name.txt."
	Convey("Given a string "+invalidFilename, t, func() {
		Convey("When the string is sanitized", func() {
			result := SanitizeFilename(invalidFilename)
			Convey("Then the result should be 'C_invalid_file_name.txt'", func() {
				So(result, ShouldEqual, "C_invalid_file_name.txt")
			})
		})
	})

	validFilename := "valid-file-name.txt"
	Convey("Given a string "+validFilename, t, func() {
		Convey("When the string is sanitized", func() {
			result := SanitizeFilename(validFilename)
			Convey("Then the result should be 'valid-file-name.txt'", func() {
				So(result, ShouldEqual, validFilename)
			})
		})
	})
}

func TestTerminalSize(t *testing.T) {
	t.Skipf("Cannot test terminal size")
}

func TestMinMax(t *testing.T) {
	t.Run("max", func(t *testing.T) {
		tests := []struct {
			name  string
			input []int
			want  int
		}{
			{name: "ordered", input: []int{1, 2, 3, 4}, want: 4},
			{name: "reversed", input: []int{9, 5, 7}, want: 9},
			{name: "negative", input: []int{-5, -1, -3}, want: -1},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := Max(tt.input...); got != tt.want {
					t.Errorf("Max() = %d, want %d", got, tt.want)
				}
			})
		}
	})

	t.Run("min", func(t *testing.T) {
		tests := []struct {
			name  string
			input []int
			want  int
		}{
			{name: "ordered", input: []int{1, 2, 3, 4}, want: 1},
			{name: "reversed", input: []int{9, 5, 7}, want: 5},
			{name: "negative", input: []int{-5, -1, -3}, want: -5},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := Min(tt.input...); got != tt.want {
					t.Errorf("Min() = %d, want %d", got, tt.want)
				}
			})
		}
	})
}

func TestReGroups(t *testing.T) {
	pattern := regexp.MustCompile(`(?P<name>[a-z]+)-(?P<number>\d+)`)

	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{name: "match", input: "foo-123", want: map[string]string{"name": "foo", "number": "123"}},
		{name: "no match", input: "!!!", want: map[string]string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReGroups(pattern, tt.input)
			for k, want := range tt.want {
				if got[k] != want {
					t.Errorf("group %q = %q, want %q", k, got[k], want)
				}
			}
		})
	}
}
