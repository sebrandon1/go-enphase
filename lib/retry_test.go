package lib

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newRetryClient() *http.Client {
	return &http.Client{Transport: &retryTransport{
		inner:     http.DefaultTransport,
		maxTries:  3,
		baseDelay: time.Millisecond,
	}}
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

	resp, err := newRetryClient().Get(srv.URL)
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
	resp, err := newRetryClient().Get(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if elapsed := time.Since(start); elapsed < time.Second {
		t.Errorf("elapsed %s, want >= 1s (Retry-After)", elapsed)
	}
}

func TestRetryExhausted429ReturnsResponse(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	resp, err := newRetryClient().Get(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests || calls.Load() != 4 {
		t.Errorf("status=%d calls=%d, want 429 and 4", resp.StatusCode, calls.Load())
	}
}

func TestRetryDoesNotRetryPost(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	resp, err := newRetryClient().Post(srv.URL, "text/plain", nil)
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
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)

	start := time.Now()
	resp, err := newRetryClient().Do(req)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected context error")
	}
	if time.Since(start) > 5*time.Second {
		t.Errorf("wait was not aborted by context cancel")
	}
}

func TestRetryAfterParsing(t *testing.T) {
	tests := []struct {
		header string
		want   time.Duration
		ok     bool
	}{
		{"5", 5 * time.Second, true},
		{"999", maxRetryDelay, true},
		{"-1", 0, false},
		{"abc", 0, false},
		{"", 0, false},
		{"0", 0, true},
	}
	for _, tt := range tests {
		resp := &http.Response{Header: http.Header{}}
		if tt.header != "" {
			resp.Header.Set("Retry-After", tt.header)
		}
		got, ok := retryAfter(resp)
		if got != tt.want || ok != tt.ok {
			t.Errorf("retryAfter(%q) = %v,%v want %v,%v", tt.header, got, ok, tt.want, tt.ok)
		}
	}
}

func TestRetryAfterNilResponse(t *testing.T) {
	if d, ok := retryAfter(nil); ok || d != 0 {
		t.Errorf("retryAfter(nil) = %v,%v want 0,false", d, ok)
	}
}

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name string
		resp *http.Response
		err  error
		want bool
	}{
		{"200 ok", &http.Response{StatusCode: http.StatusOK}, nil, false},
		{"404 not found", &http.Response{StatusCode: http.StatusNotFound}, nil, false},
		{"429 too many requests", &http.Response{StatusCode: http.StatusTooManyRequests}, nil, true},
		{"500 server error", &http.Response{StatusCode: http.StatusInternalServerError}, nil, true},
		{"503 unavailable", &http.Response{StatusCode: http.StatusServiceUnavailable}, nil, true},
		{"network timeout", nil, timeoutErr{}, true},
		{"context canceled", nil, context.Canceled, false},
		{"deadline exceeded", nil, context.DeadlineExceeded, false},
		{"non-timeout error", nil, errors.New("connection refused"), false},
		{"nil response and error", nil, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRetry(tt.resp, tt.err); got != tt.want {
				t.Errorf("shouldRetry() = %v, want %v", got, tt.want)
			}
		})
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

	resp, err := newRetryClient().Get(srv.URL)
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

	resp, err := newRetryClient().Get(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if calls.Load() != 1 {
		t.Errorf("calls=%d, want 1", calls.Load())
	}
}
