package main

import (
	"net/http"
	"time"
)

func healthzHandler(col *Collector, pollInterval time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := col.Snapshot()
		if snap.UpdatedAt.IsZero() || time.Since(snap.UpdatedAt) > 2*pollInterval {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("stale: no successful collection within 2x poll interval")) //nolint:errcheck
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok")) //nolint:errcheck
	}
}
