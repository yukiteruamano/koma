package weebcentral

import "testing"

func TestAbsoluteURL(t *testing.T) {
	tests := []struct {
		name string
		href string
		want string
	}{
		{name: "relative path", href: "/chapters/01ABC", want: "https://weebcentral.com/chapters/01ABC"},
		{name: "absolute url", href: "https://weebcentral.com/series/01XYZ", want: "https://weebcentral.com/series/01XYZ"},
		{name: "protocol-relative", href: "//cdn.weebcentral.com/img.jpg", want: "https://cdn.weebcentral.com/img.jpg"},
		{name: "empty", href: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := absoluteURL(tt.href); got != tt.want {
				t.Errorf("absoluteURL(%q) = %q, want %q", tt.href, got, tt.want)
			}
		})
	}
}

func TestURLID(t *testing.T) {
	tests := []struct {
		name string
		href string
		want string
	}{
		{name: "absolute url", href: "https://weebcentral.com/chapters/01ABC", want: "01ABC"},
		{name: "relative url", href: "/chapters/01ABC", want: "01ABC"},
		{name: "url with trailing slash", href: "https://weebcentral.com/series/01XYZ/", want: "01XYZ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := urlID(tt.href); got != tt.want {
				t.Errorf("urlID(%q) = %q, want %q", tt.href, got, tt.want)
			}
		})
	}
}
