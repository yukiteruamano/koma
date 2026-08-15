package network

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/key"
)

func setRetryConfig(t *testing.T) {
	t.Helper()
	viper.Set(key.NetworkMaxRetries, 3)
	viper.Set(key.NetworkRetryBaseDelay, 5*time.Millisecond)
}

func TestDoRetriesTransientResponses(t *testing.T) {
	setRetryConfig(t)

	tests := []struct {
		name         string
		statuses     []int
		withHeader   bool
		wantAttempts int
		wantErr      bool
		wantStatusOK bool
	}{
		{
			name:         "500 then success",
			statuses:     []int{http.StatusInternalServerError, http.StatusInternalServerError, http.StatusOK},
			wantAttempts: 3,
			wantStatusOK: true,
		},
		{
			name:         "429 then success",
			statuses:     []int{http.StatusTooManyRequests, http.StatusTooManyRequests, http.StatusOK},
			wantAttempts: 3,
			wantStatusOK: true,
		},
		{
			name:         "honors retry-after header",
			statuses:     []int{http.StatusTooManyRequests, http.StatusOK},
			withHeader:   true,
			wantAttempts: 2,
			wantStatusOK: true,
		},
		{
			name:         "exhausts retries",
			statuses:     []int{500, 500, 500, 500},
			wantAttempts: 4,
			wantErr:      true,
		},
		{
			name:         "does not retry 404",
			statuses:     []int{http.StatusNotFound},
			wantAttempts: 1,
			wantStatusOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				n := int(attempts.Add(1)) - 1
				if tt.withHeader && tt.statuses[n] == http.StatusTooManyRequests {
					w.Header().Set("Retry-After", "0")
				}
				w.WriteHeader(tt.statuses[n])
			}))
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL, nil)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := Do(req)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if int(attempts.Load()) != tt.wantAttempts {
				t.Errorf("attempts = %d, want %d", attempts.Load(), tt.wantAttempts)
			}

			if tt.wantStatusOK && (resp == nil || resp.StatusCode != http.StatusOK) {
				t.Errorf("expected 200 response, got %+v", resp)
			}

			if resp != nil {
				_ = resp.Body.Close()
			}
		})
	}
}

func TestBackoffDelayClamped(t *testing.T) {
	base := 5 * time.Millisecond

	for _, attempt := range []int{1, 5, 20, 40, 60, 70} {
		t.Run("attempt", func(t *testing.T) {
			delay := backoffDelay(base, attempt)
			if delay < 0 {
				t.Fatalf("negative delay for attempt %d: %v", attempt, delay)
			}
			if delay > maxBackoff {
				t.Fatalf("delay %v exceeds maxBackoff %v", delay, maxBackoff)
			}
		})
	}
}

func TestBackoffDelayFirstAttempt(t *testing.T) {
	base := 10 * time.Millisecond
	delay := backoffDelay(base, 1)
	if delay < base/2 || delay > base {
		t.Errorf("first-attempt delay %v not in [%v, %v]", delay, base/2, base)
	}
}

func TestDoNegativeRetriesDisablesRetry(t *testing.T) {
	viper.Set(key.NetworkMaxRetries, -1)
	viper.Set(key.NetworkRetryBaseDelay, 5*time.Millisecond)

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	if _, err := Do(req); err == nil {
		t.Fatal("expected an error after exhausting retries")
	}

	if attempts.Load() != 1 {
		t.Errorf("attempts = %d, want 1 (negative retries must disable retry)", attempts.Load())
	}
}

func TestDoZeroBaseDelayDoesNotPanic(t *testing.T) {
	viper.Set(key.NetworkMaxRetries, 2)
	viper.Set(key.NetworkRetryBaseDelay, 0)

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if attempts.Load() != 3 {
		t.Errorf("attempts = %d, want 3", attempts.Load())
	}
}

func TestDoNoRetryOnNonRetryableError(t *testing.T) {
	setRetryConfig(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// a body that cannot be recreated disables retries
	body := &nopCloser{}
	req, err := http.NewRequest(http.MethodPost, server.URL, body)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Do(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type nopCloser struct{}

func (nopCloser) Read([]byte) (int, error) { return 0, io.EOF }
func (nopCloser) Close() error             { return nil }

func TestIsTransient(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "net timeout",
			err:  net.Error(netTimeoutError{}),
			want: true,
		},
		{
			name: "plain error",
			err:  errors.New("boom"),
			want: false,
		},
		{
			name: "nil",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTransient(tt.err); got != tt.want {
				t.Errorf("isTransient(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

type netTimeoutError struct{}

func (netTimeoutError) Error() string   { return "timeout" }
func (netTimeoutError) Timeout() bool   { return true }
func (netTimeoutError) Temporary() bool { return true }
