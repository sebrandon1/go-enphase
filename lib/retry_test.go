package lib

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sebrandon1/go-toolbox/httpx"
	"github.com/sebrandon1/go-toolbox/retry"
)

func newRetryClient() *http.Client {
	return &http.Client{Transport: http.DefaultTransport}
}

func doRetry(t *testing.T, client *http.Client, req *http.Request) (*http.Response, error) {
	t.Helper()
	return retry.Do(req.Context(), client, req, retryOptions(time.Millisecond)...)
}

func newGet(t *testing.T, url string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func TestRetryOn429ThenSuccess(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	resp, err := doRetry(t, newRetryClient(), newGet(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK || calls.Load() != 2 {
		t.Errorf("status=%d calls=%d, want 200 and 2", resp.StatusCode, calls.Load())
	}
}

func TestRetryHonorsRetryAfter(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	start := time.Now()
	resp, err := doRetry(t, newRetryClient(), newGet(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if elapsed := time.Since(start); elapsed < time.Second {
		t.Errorf("elapsed %s, want >= 1s (Retry-After)", elapsed)
	}
}

func TestRetryRetryAfterCappedAtMaxDelay(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "999")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	var got time.Duration
	opts := append(retryOptions(time.Millisecond),
		retry.WithHTTPOnRetry(func(attempt int, err error, delay time.Duration) {
			got = delay
			cancel() // abort the actual wait
		}),
	)
	if _, err := retry.Do(ctx, newRetryClient(), req, opts...); err == nil {
		t.Fatal("expected context error")
	}
	if got != maxRetryDelay {
		t.Errorf("retry delay = %s, want %s", got, maxRetryDelay)
	}
}

func TestRetryExhausted429ReturnsTypedError(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	resp, err := doRetry(t, newRetryClient(), newGet(t, srv.URL))
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected error after retries are exhausted")
	}
	var rateErr *httpx.RateLimitError
	if !errors.As(err, &rateErr) {
		t.Errorf("error %v is not a RateLimitError", err)
	}
	if calls.Load() != maxRetryAttempts {
		t.Errorf("calls=%d, want %d", calls.Load(), maxRetryAttempts)
	}
}

func TestRetryDoesNotRetryPost(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := doRetry(t, newRetryClient(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if calls.Load() != 1 {
		t.Errorf("calls=%d, want 1", calls.Load())
	}
}

func TestRetryContextCancelAbortsWait(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	resp, err := doRetry(t, newRetryClient(), req)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected context error")
	}
	if time.Since(start) > 5*time.Second {
		t.Errorf("wait was not aborted by context cancel")
	}
}

func TestRetryOn5xxThenSuccess(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	resp, err := doRetry(t, newRetryClient(), newGet(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK || calls.Load() != 2 {
		t.Errorf("status=%d calls=%d, want 200 and 2", resp.StatusCode, calls.Load())
	}
}

func TestNoRetryOn4xx(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	resp, err := doRetry(t, newRetryClient(), newGet(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if calls.Load() != 1 {
		t.Errorf("calls=%d, want 1", calls.Load())
	}
}

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{http.StatusOK, false},
		{http.StatusNotFound, false},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusBadGateway, true},
	}
	for _, tt := range tests {
		if got := isRetryableStatus(tt.code); got != tt.want {
			t.Errorf("isRetryableStatus(%d) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestDoGetConvertsRetryExhaustionToFriendlyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := &Client{HTTPClient: newRetryClient()}
	resp, err := c.doGet(context.Background(), newGet(t, srv.URL))
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "rate limit exceeded (429): wait before retrying" {
		t.Errorf("error = %q, want rate limit message", err.Error())
	}
	if !errors.As(err, new(*httpx.RateLimitError)) {
		t.Errorf("error %v does not wrap RateLimitError", err)
	}
}

func TestDoGetReturnsNonRetryableResponse(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := &Client{HTTPClient: newRetryClient()}
	resp, err := c.doGet(context.Background(), newGet(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized || calls.Load() != 1 {
		t.Errorf("status=%d calls=%d, want 401 and 1", resp.StatusCode, calls.Load())
	}
}
