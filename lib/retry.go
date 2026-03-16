package lib

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// retryTransport is an http.RoundTripper that retries GET requests on 5xx responses
// and transient network errors with exponential backoff.
type retryTransport struct {
	inner     http.RoundTripper
	maxTries  int
	baseDelay time.Duration
}

// RoundTrip implements http.RoundTripper with retry logic for GET requests.
func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodGet {
		return rt.inner.RoundTrip(req)
	}

	var (
		resp  *http.Response
		err   error
		delay = rt.baseDelay
	)

	for attempt := 0; attempt <= rt.maxTries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(delay):
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
			delay = min(delay*2, 30*time.Second)
		}

		resp, err = rt.inner.RoundTrip(req)
		if !shouldRetry(resp, err) {
			return resp, err
		}
		if resp != nil {
			drainAndClose(resp.Body)
		}
	}

	return resp, err
}

// shouldRetry reports whether a response or error warrants a retry.
func shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}
		var netErr net.Error
		return errors.As(err, &netErr) && netErr.Timeout()
	}
	return resp != nil && resp.StatusCode >= 500
}
