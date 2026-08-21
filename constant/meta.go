package constant

import "math/rand"

const (
	Koma    = "koma"
	Version = "1.0.0"
)

// UserAgent is kept for backwards compatibility; use RandomUserAgent() for rotation.
const UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36"

// UserAgents contains at least 10 up-to-date User-Agent strings for Linux, macOS and Android.
// Chrome 151/150/143, Firefox 153, Safari 26 — verified Aug 2026 via Exa MCP (UA reduction: 0.0.0, Android 10; K, macOS 10_15_7).
var UserAgents = []string{
	// Chrome — Linux / macOS / Android (Chrome 151, current stable)
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Mobile Safari/537.36",
	// Chrome — previous stable 150 (still valid due to frozen 0.0.0)
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Mobile Safari/537.36",
	// Firefox 153 — Linux / macOS / Android
	"Mozilla/5.0 (X11; Linux x86_64; rv:153.0) Gecko/20100101 Firefox/153.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:153.0) Gecko/20100101 Firefox/153.0",
	"Mozilla/5.0 (Android 10; Mobile; rv:153.0) Gecko/153.0 Firefox/153.0",
	// Safari 26 — macOS
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/26.0 Safari/605.1.15",
	// Chrome 143 — fallback/rotation diversity
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
}

// RandomUserAgent returns a random User-Agent from the pool.
func RandomUserAgent() string {
	return UserAgents[rand.Intn(len(UserAgents))]
}
