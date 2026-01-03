package app

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// DiagnosticsState holds diagnostic metrics for the debug overlay.
type DiagnosticsState struct {
	// Request timing
	LastRequestLatency time.Duration
	RequestCount       int
	TotalRequestTime   time.Duration

	// Cache stats
	ArtworkCacheHits   int
	ArtworkCacheMisses int

	// Player state
	MPVConnected   bool
	MPVReconnects  int
	LastMPVError   string
	LastMPVErrorAt time.Time

	// Visualizer
	VisualizerRunning bool
	VisualizerFPS     int

	// App stats
	StartTime      time.Time
	LastUpdate     time.Time
	MemoryUsage    uint64
	GoroutineCount int
}

// NewDiagnosticsState creates a new diagnostics state.
func NewDiagnosticsState() *DiagnosticsState {
	return &DiagnosticsState{
		StartTime:    time.Now(),
		MPVConnected: true,
	}
}

// RecordRequest records a provider request latency.
func (d *DiagnosticsState) RecordRequest(latency time.Duration) {
	d.LastRequestLatency = latency
	d.RequestCount++
	d.TotalRequestTime += latency
}

// AverageLatency returns the average request latency.
func (d *DiagnosticsState) AverageLatency() time.Duration {
	if d.RequestCount == 0 {
		return 0
	}
	return d.TotalRequestTime / time.Duration(d.RequestCount)
}

// RecordArtworkCacheHit records an artwork cache hit.
func (d *DiagnosticsState) RecordArtworkCacheHit() {
	d.ArtworkCacheHits++
}

// RecordArtworkCacheMiss records an artwork cache miss.
func (d *DiagnosticsState) RecordArtworkCacheMiss() {
	d.ArtworkCacheMisses++
}

// ArtworkCacheHitRate returns the cache hit rate as a percentage.
func (d *DiagnosticsState) ArtworkCacheHitRate() float64 {
	total := d.ArtworkCacheHits + d.ArtworkCacheMisses
	if total == 0 {
		return 0
	}
	return float64(d.ArtworkCacheHits) / float64(total) * 100
}

// RecordMPVError records an mpv error.
func (d *DiagnosticsState) RecordMPVError(err string) {
	d.LastMPVError = err
	d.LastMPVErrorAt = time.Now()
	d.MPVConnected = false
}

// RecordMPVReconnect records an mpv reconnection.
func (d *DiagnosticsState) RecordMPVReconnect() {
	d.MPVReconnects++
	d.MPVConnected = true
}

// Update refreshes runtime stats.
func (d *DiagnosticsState) Update() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	d.MemoryUsage = m.Alloc
	d.GoroutineCount = runtime.NumGoroutine()
	d.LastUpdate = time.Now()
}

// Uptime returns the application uptime.
func (d *DiagnosticsState) Uptime() time.Duration {
	return time.Since(d.StartTime)
}

// Render renders the diagnostics overlay.
func (d *DiagnosticsState) Render(m *Model) string {
	d.Update()

	var b strings.Builder

	// Header
	b.WriteString(m.theme.Title.Render(" ═══ Diagnostics ═══ "))
	b.WriteString("\n\n")

	// Uptime
	uptime := d.Uptime().Round(time.Second)
	b.WriteString(m.theme.Dim.Render("Uptime: "))
	b.WriteString(m.theme.Text.Render(uptime.String()))
	b.WriteString("\n\n")

	// Memory & Goroutines
	b.WriteString(m.theme.Accent.Render("Runtime"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Memory: %s\n", formatBytes(d.MemoryUsage)))
	b.WriteString(fmt.Sprintf("  Goroutines: %d\n", d.GoroutineCount))
	b.WriteString("\n")

	// Provider requests
	b.WriteString(m.theme.Accent.Render("Provider"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Requests: %d\n", d.RequestCount))
	if d.RequestCount > 0 {
		b.WriteString(fmt.Sprintf("  Last latency: %s\n", d.LastRequestLatency.Round(time.Millisecond)))
		b.WriteString(fmt.Sprintf("  Avg latency: %s\n", d.AverageLatency().Round(time.Millisecond)))
	}
	b.WriteString("\n")

	// Artwork cache
	b.WriteString(m.theme.Accent.Render("Artwork Cache"))
	b.WriteString("\n")
	total := d.ArtworkCacheHits + d.ArtworkCacheMisses
	if total > 0 {
		b.WriteString(fmt.Sprintf("  Hits: %d / Misses: %d\n", d.ArtworkCacheHits, d.ArtworkCacheMisses))
		b.WriteString(fmt.Sprintf("  Hit rate: %.1f%%\n", d.ArtworkCacheHitRate()))
	} else {
		b.WriteString("  No requests yet\n")
	}
	b.WriteString("\n")

	// mpv status
	b.WriteString(m.theme.Accent.Render("mpv Player"))
	b.WriteString("\n")
	if d.MPVConnected {
		b.WriteString(m.theme.Success.Render("  ● Connected"))
	} else {
		b.WriteString(m.theme.Error.Render("  ○ Disconnected"))
	}
	b.WriteString("\n")
	if d.MPVReconnects > 0 {
		b.WriteString(fmt.Sprintf("  Reconnects: %d\n", d.MPVReconnects))
	}
	if d.LastMPVError != "" && time.Since(d.LastMPVErrorAt) < 5*time.Minute {
		b.WriteString(m.theme.Error.Render(fmt.Sprintf("  Last error: %s\n", d.LastMPVError)))
	}
	b.WriteString("\n")

	// Visualizer
	b.WriteString(m.theme.Accent.Render("Visualizer"))
	b.WriteString("\n")
	if m.visualizer != nil && d.VisualizerRunning {
		b.WriteString(m.theme.Success.Render("  ● Running"))
		b.WriteString(fmt.Sprintf(" (~%d fps)\n", d.VisualizerFPS))
	} else if m.visualizer != nil {
		b.WriteString(m.theme.Dim.Render("  ○ Stopped\n"))
	} else {
		b.WriteString(m.theme.Dim.Render("  ○ Not available (cava not installed)\n"))
	}
	b.WriteString("\n")

	// Playback
	b.WriteString(m.theme.Accent.Render("Playback"))
	b.WriteString("\n")
	if m.nowPlaying.Title != "" {
		state := "Playing"
		if m.paused {
			state = "Paused"
		}
		b.WriteString(fmt.Sprintf("  State: %s\n", state))
		b.WriteString(fmt.Sprintf("  Volume: %.0f%% %s\n", m.volume, map[bool]string{true: "(muted)", false: ""}[m.muted]))
		b.WriteString(fmt.Sprintf("  Position: %.0f / %.0f sec\n", m.timePos, m.duration))
	} else {
		b.WriteString("  Nothing playing\n")
	}
	b.WriteString("\n")

	// Queue
	b.WriteString(m.theme.Accent.Render("Queue"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Items: %d\n", m.queue.Len()))
	b.WriteString(fmt.Sprintf("  Current: %d\n", m.queue.CurrentIndex()))
	b.WriteString(fmt.Sprintf("  Shuffle: %v\n", m.queue.IsShuffled()))
	b.WriteString(fmt.Sprintf("  Repeat: %v\n", m.queue.RepeatMode()))

	// Footer
	b.WriteString("\n")
	b.WriteString(m.theme.Dim.Render("Press Ctrl+D to close"))

	// Wrap in a box
	content := b.String()
	diagBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(40).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Right, lipgloss.Top, diagBox)
}

// formatBytes formats bytes as human-readable string.
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
