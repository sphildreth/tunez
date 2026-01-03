// Package visualizer provides real-time audio spectrum visualization using CAVA.
package visualizer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Visualizer manages CAVA subprocess for audio spectrum analysis.
type Visualizer struct {
	mu       sync.RWMutex
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	bars     []int
	barCount int
	maxValue int
	running  bool
	err      error
}

// Config holds visualizer configuration.
type Config struct {
	BarCount int // Number of frequency bars (default: 24)
	MaxValue int // Maximum bar value for scaling (default: 1000)
}

// New creates a new Visualizer instance.
func New(cfg Config) *Visualizer {
	if cfg.BarCount <= 0 {
		cfg.BarCount = 24 // Wider default
	}
	if cfg.MaxValue <= 0 {
		cfg.MaxValue = 1000
	}
	return &Visualizer{
		barCount: cfg.BarCount,
		maxValue: cfg.MaxValue,
		bars:     make([]int, cfg.BarCount),
	}
}

// Available checks if CAVA is installed on the system.
func Available() bool {
	_, err := exec.LookPath("cava")
	return err == nil
}

// Start begins the CAVA subprocess and starts reading spectrum data.
func (v *Visualizer) Start(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.running {
		return nil
	}

	if !Available() {
		v.err = fmt.Errorf("cava not installed")
		return v.err
	}

	// Create temporary config file for CAVA
	configPath, err := v.writeConfig()
	if err != nil {
		v.err = err
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	v.cancel = cancel

	v.cmd = exec.CommandContext(ctx, "cava", "-p", configPath)
	v.cmd.Stderr = nil // Suppress stderr

	stdout, err := v.cmd.StdoutPipe()
	if err != nil {
		v.err = err
		return err
	}

	if err := v.cmd.Start(); err != nil {
		v.err = err
		return err
	}

	v.running = true
	v.err = nil

	// Read CAVA output in background
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				v.mu.Lock()
				v.running = false
				v.mu.Unlock()
				return
			default:
				line := scanner.Text()
				v.parseLine(line)
			}
		}
		v.mu.Lock()
		v.running = false
		v.mu.Unlock()
	}()

	// Cleanup config file when done
	go func() {
		<-ctx.Done()
		os.Remove(configPath)
	}()

	return nil
}

// Stop terminates the CAVA subprocess.
func (v *Visualizer) Stop() {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.cancel != nil {
		v.cancel()
		v.cancel = nil
	}
	if v.cmd != nil && v.cmd.Process != nil {
		v.cmd.Process.Kill()
		v.cmd.Wait()
		v.cmd = nil
	}
	v.running = false
}

// Bars returns a copy of the current bar values (0 to maxValue).
func (v *Visualizer) Bars() []int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	result := make([]int, len(v.bars))
	copy(result, v.bars)
	return result
}

// BarsNormalized returns bar values normalized to 0-8 range for display.
func (v *Visualizer) BarsNormalized() []int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	result := make([]int, len(v.bars))
	for i, val := range v.bars {
		// Normalize to 0-8 range
		normalized := (val * 8) / v.maxValue
		if normalized > 8 {
			normalized = 8
		}
		result[i] = normalized
	}
	return result
}

// Running returns true if the visualizer is active.
func (v *Visualizer) Running() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.running
}

// Error returns any error that occurred.
func (v *Visualizer) Error() error {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.err
}

// parseLine parses a line of CAVA raw output.
func (v *Visualizer) parseLine(line string) {
	parts := strings.Split(line, ";")
	v.mu.Lock()
	defer v.mu.Unlock()

	for i := 0; i < len(v.bars) && i < len(parts); i++ {
		if val, err := strconv.Atoi(strings.TrimSpace(parts[i])); err == nil {
			v.bars[i] = val
		}
	}
}

// writeConfig creates a temporary CAVA config file.
func (v *Visualizer) writeConfig() (string, error) {
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, fmt.Sprintf("tunez-cava-%d.conf", os.Getpid()))

	config := fmt.Sprintf(`[general]
bars = %d
framerate = 30
sensitivity = 100
autosens = 1
overshoot = 20

[input]
method = pulse
; Falls back to other methods automatically

[output]
method = raw
data_format = ascii
ascii_max_range = %d
bar_delimiter = 59
frame_delimiter = 10

[smoothing]
integral = 77
monstercat = 0
waves = 0
gravity = 100

[eq]
; Equal weight across bands
`, v.barCount, v.maxValue)

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return "", fmt.Errorf("write cava config: %w", err)
	}

	return configPath, nil
}

// Render returns an ANSI string representation of the current spectrum.
// Uses Unicode block characters for smooth visualization.
func (v *Visualizer) Render() string {
	bars := v.BarsNormalized()
	if len(bars) == 0 {
		return ""
	}

	// Unicode block characters for different heights
	blocks := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	var sb strings.Builder
	sb.WriteString("║")
	for _, val := range bars {
		if val < 0 {
			val = 0
		}
		if val > 8 {
			val = 8
		}
		sb.WriteRune(blocks[val])
	}
	sb.WriteString("║")

	return sb.String()
}

// RenderRainbow returns a rainbow-colored visualization using full blocks.
// Each bar gets a color from the rainbow spectrum based on its position.
func (v *Visualizer) RenderRainbow() string {
	bars := v.BarsNormalized()
	if len(bars) == 0 {
		return ""
	}

	// Rainbow colors (ANSI 256-color codes)
	rainbowColors := []int{196, 202, 208, 214, 220, 226, 190, 154, 118, 82, 46, 47, 48, 49, 50, 51, 45, 39, 33, 27, 21, 57, 93, 129}

	// Unicode block characters for different heights
	blocks := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	var sb strings.Builder
	sb.WriteString("║")
	for i, val := range bars {
		if val < 0 {
			val = 0
		}
		if val > 8 {
			val = 8
		}
		// Pick color based on bar position
		colorIdx := (i * len(rainbowColors)) / len(bars)
		if colorIdx >= len(rainbowColors) {
			colorIdx = len(rainbowColors) - 1
		}
		color := rainbowColors[colorIdx]

		if val == 0 {
			sb.WriteString(" ")
		} else {
			sb.WriteString(fmt.Sprintf("\x1b[38;5;%dm%c\x1b[0m", color, blocks[val]))
		}
	}
	sb.WriteString("║")

	return sb.String()
}

// RenderTall returns a multi-line visualization for taller display.
// height specifies how many terminal rows to use.
// Deprecated: Use RenderSized instead.
func (v *Visualizer) RenderTall(height int, rainbow bool) string {
	return v.RenderSized(0, height, rainbow)
}

// RenderSized returns a visualization with specified width and height.
// width is the number of characters (0 = use bar count)
// height is the number of terminal rows (0 = auto based on width)
func (v *Visualizer) RenderSized(width, height int, rainbow bool) string {
	bars := v.Bars()
	if len(bars) == 0 {
		return ""
	}

	// Default width to bar count
	if width <= 0 {
		width = len(bars)
	}

	// Auto height: roughly 1 row per 12-15 chars width for good aspect ratio
	if height <= 0 {
		height = (width + 11) / 12
		if height < 2 {
			height = 2
		}
		if height > 6 {
			height = 6
		}
	}

	// Rainbow colors (ANSI 256-color codes) - full spectrum
	rainbowColors := []int{196, 202, 208, 214, 220, 226, 190, 154, 118, 82, 46, 47, 48, 49, 50, 51, 45, 39, 33, 27, 21, 57, 93, 129}

	// Interpolate/stretch bars to fit width
	stretched := make([]int, width)
	for i := 0; i < width; i++ {
		// Map output position to input bar
		srcIdx := (i * len(bars)) / width
		if srcIdx >= len(bars) {
			srcIdx = len(bars) - 1
		}
		stretched[i] = bars[srcIdx]
	}

	// Normalize to height * 8 for sub-character resolution
	maxHeight := height * 8
	normalized := make([]int, width)
	for i, val := range stretched {
		n := (val * maxHeight) / v.maxValue
		if n > maxHeight {
			n = maxHeight
		}
		normalized[i] = n
	}

	// Build from top to bottom
	var lines []string
	blocks := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	for row := height - 1; row >= 0; row-- {
		var sb strings.Builder
		sb.WriteString("║")
		threshold := row * 8

		for i, val := range normalized {
			remaining := val - threshold
			if remaining <= 0 {
				sb.WriteString(" ")
			} else {
				blockIdx := remaining
				if blockIdx > 8 {
					blockIdx = 8
				}

				if rainbow {
					colorIdx := (i * len(rainbowColors)) / width
					if colorIdx >= len(rainbowColors) {
						colorIdx = len(rainbowColors) - 1
					}
					color := rainbowColors[colorIdx]
					sb.WriteString(fmt.Sprintf("\x1b[38;5;%dm%c\x1b[0m", color, blocks[blockIdx]))
				} else {
					sb.WriteRune(blocks[blockIdx])
				}
			}
		}
		sb.WriteString("║")
		lines = append(lines, sb.String())
	}

	return strings.Join(lines, "\n")
}
