package app

import (
	"testing"
	"time"
)

func TestDiagnosticsState(t *testing.T) {
	d := NewDiagnosticsState()

	t.Run("initial state", func(t *testing.T) {
		if d.RequestCount != 0 {
			t.Errorf("expected 0 requests, got %d", d.RequestCount)
		}
		if !d.MPVConnected {
			t.Error("expected mpv to be connected initially")
		}
		if d.StartTime.IsZero() {
			t.Error("expected start time to be set")
		}
	})

	t.Run("record request", func(t *testing.T) {
		d.RecordRequest(100 * time.Millisecond)
		if d.RequestCount != 1 {
			t.Errorf("expected 1 request, got %d", d.RequestCount)
		}
		if d.LastRequestLatency != 100*time.Millisecond {
			t.Errorf("expected 100ms, got %v", d.LastRequestLatency)
		}
	})

	t.Run("average latency", func(t *testing.T) {
		d2 := NewDiagnosticsState()
		d2.RecordRequest(100 * time.Millisecond)
		d2.RecordRequest(200 * time.Millisecond)
		avg := d2.AverageLatency()
		if avg != 150*time.Millisecond {
			t.Errorf("expected 150ms, got %v", avg)
		}
	})

	t.Run("cache hit rate", func(t *testing.T) {
		d3 := NewDiagnosticsState()
		d3.RecordArtworkCacheHit()
		d3.RecordArtworkCacheHit()
		d3.RecordArtworkCacheMiss()
		rate := d3.ArtworkCacheHitRate()
		expected := 66.66666666666667
		if rate < 66.6 || rate > 66.7 {
			t.Errorf("expected ~%.2f%%, got %.2f%%", expected, rate)
		}
	})

	t.Run("mpv error tracking", func(t *testing.T) {
		d4 := NewDiagnosticsState()
		d4.RecordMPVError("connection lost")
		if d4.MPVConnected {
			t.Error("expected mpv to be disconnected after error")
		}
		if d4.LastMPVError != "connection lost" {
			t.Errorf("expected 'connection lost', got '%s'", d4.LastMPVError)
		}

		d4.RecordMPVReconnect()
		if !d4.MPVConnected {
			t.Error("expected mpv to be connected after reconnect")
		}
		if d4.MPVReconnects != 1 {
			t.Errorf("expected 1 reconnect, got %d", d4.MPVReconnects)
		}
	})

	t.Run("update refreshes stats", func(t *testing.T) {
		d5 := NewDiagnosticsState()
		d5.Update()
		if d5.MemoryUsage == 0 {
			t.Error("expected memory usage to be set")
		}
		if d5.GoroutineCount == 0 {
			t.Error("expected goroutine count to be set")
		}
		if d5.LastUpdate.IsZero() {
			t.Error("expected last update to be set")
		}
	})

	t.Run("uptime increases", func(t *testing.T) {
		d6 := NewDiagnosticsState()
		time.Sleep(10 * time.Millisecond)
		uptime := d6.Uptime()
		if uptime < 10*time.Millisecond {
			t.Errorf("expected uptime > 10ms, got %v", uptime)
		}
	})
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatBytes(tc.bytes)
			if result != tc.expected {
				t.Errorf("formatBytes(%d) = %s, want %s", tc.bytes, result, tc.expected)
			}
		})
	}
}
