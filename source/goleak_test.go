package source

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// httptest servers + the shared HTTP client can leave http2
		// read-loop goroutines behind after a test finishes.
		goleak.IgnoreAnyFunction("net/http.(*http2clientConnReadLoop).run"),
	)
}
