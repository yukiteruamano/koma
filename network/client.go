package network

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/key"
)

var transport = http.DefaultTransport.(*http.Transport).Clone()

func init() {
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 100
	transport.IdleConnTimeout = 30 * time.Second
	transport.ResponseHeaderTimeout = 30 * time.Second
	transport.ExpectContinueTimeout = 30 * time.Second
}

var (
	clientOnce sync.Once
	client     *http.Client
)

// Client returns a lazily built http.Client configured from viper defaults.
// The transport is shared, so tuning happens once on first use.
func Client() *http.Client {
	clientOnce.Do(func() {
		if maxConns := viper.GetInt(key.NetworkMaxConnsPerHost); maxConns > 0 {
			transport.MaxConnsPerHost = maxConns
		}

		timeout := viper.GetDuration(key.NetworkTimeout)
		if timeout <= 0 {
			timeout = time.Minute
		}

		client = &http.Client{Timeout: timeout, Transport: transport}
	})

	return client
}

// Do performs a request, retrying transient failures with exponential backoff
// and jitter. Only requests with a nil or re-creatable body are retried.
func Do(req *http.Request) (*http.Response, error) {
	maxRetries := viper.GetInt(key.NetworkMaxRetries)
	if maxRetries < 0 {
		maxRetries = 0
	}

	baseDelay := viper.GetDuration(key.NetworkRetryBaseDelay)
	if baseDelay <= 0 {
		baseDelay = 500 * time.Millisecond
	}

	// Requests with a body that cannot be recreated are not retried.
	retryable := req.Body == nil || req.GetBody != nil

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 && retryable {
			delay := baseDelay << (attempt - 1)
			delay = delay/2 + time.Duration(rand.Int63n(int64(delay)/2+1))
			time.Sleep(delay)
		}

		var reqAttempt *http.Request
		if attempt == 0 || !retryable {
			reqAttempt = req
		} else {
			var err error
			reqAttempt, err = cloneRequest(req)
			if err != nil {
				return nil, err
			}
		}

		resp, err := Client().Do(reqAttempt)
		if err != nil {
			if retryable && isTransient(err) {
				lastErr = err
				continue
			}
			return resp, err
		}

		// Retry rate-limited and server-error responses.
		if retryable && (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) {
			_ = resp.Body.Close()
			lastErr = errors.New("http error: " + resp.Status)
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}

// cloneRequest returns a copy of req suitable for resending.
func cloneRequest(req *http.Request) (*http.Request, error) {
	if req.GetBody == nil {
		return req.Clone(context.Background()), nil
	}

	body, err := req.GetBody()
	if err != nil {
		return nil, err
	}

	clone := req.Clone(context.Background())
	clone.Body = body
	return clone, nil
}

func isTransient(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}

	return false
}
