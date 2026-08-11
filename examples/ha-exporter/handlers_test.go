package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthzNeverCollected(t *testing.T) {
	col := &Collector{}
	h := healthzHandler(col, 30*time.Second)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when UpdatedAt is zero, got %d", w.Code)
	}
	if w.Body.String() == "ok" {
		t.Error("expected stale body, got 'ok'")
	}
}

func TestHealthzStale(t *testing.T) {
	col := &Collector{}
	col.mu.Lock()
	col.snap.UpdatedAt = time.Now().Add(-3 * time.Minute)
	col.mu.Unlock()

	h := healthzHandler(col, 30*time.Second)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when last collection is older than 2x poll interval, got %d", w.Code)
	}
}

func TestHealthzFresh(t *testing.T) {
	col := &Collector{}
	col.mu.Lock()
	col.snap.UpdatedAt = time.Now()
	col.mu.Unlock()

	h := healthzHandler(col, 30*time.Second)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for fresh snapshot, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", w.Body.String())
	}
}
