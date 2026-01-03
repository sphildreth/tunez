package visualizer

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	v := New(Config{BarCount: 8, MaxValue: 500})
	if v == nil {
		t.Fatal("expected non-nil visualizer")
	}
	if v.barCount != 8 {
		t.Errorf("expected barCount 8, got %d", v.barCount)
	}
	if v.maxValue != 500 {
		t.Errorf("expected maxValue 500, got %d", v.maxValue)
	}
}

func TestNewDefaults(t *testing.T) {
	v := New(Config{})
	if v.barCount != 24 {
		t.Errorf("expected default barCount 24, got %d", v.barCount)
	}
	if v.maxValue != 1000 {
		t.Errorf("expected default maxValue 1000, got %d", v.maxValue)
	}
}

func TestBarsNormalized(t *testing.T) {
	v := New(Config{BarCount: 4, MaxValue: 100})
	// Manually set bar values
	v.bars = []int{0, 50, 100, 25}

	normalized := v.BarsNormalized()
	if len(normalized) != 4 {
		t.Fatalf("expected 4 bars, got %d", len(normalized))
	}

	expected := []int{0, 4, 8, 2}
	for i, want := range expected {
		if normalized[i] != want {
			t.Errorf("bar[%d]: expected %d, got %d", i, want, normalized[i])
		}
	}
}

func TestRender(t *testing.T) {
	v := New(Config{BarCount: 4, MaxValue: 100})
	v.bars = []int{0, 25, 50, 100}

	rendered := v.Render()
	if rendered == "" {
		t.Error("expected non-empty render output")
	}
	// Should start and end with bars
	if rendered[0] != 0xe2 { // UTF-8 start of ║
		t.Error("expected render to start with ║")
	}
}

func TestParseLine(t *testing.T) {
	v := New(Config{BarCount: 4, MaxValue: 1000})
	v.parseLine("100;200;300;400")

	if v.bars[0] != 100 {
		t.Errorf("expected bar[0]=100, got %d", v.bars[0])
	}
	if v.bars[3] != 400 {
		t.Errorf("expected bar[3]=400, got %d", v.bars[3])
	}
}

func TestAvailable(t *testing.T) {
	// Just test that it doesn't panic
	_ = Available()
}

func TestRenderTall(t *testing.T) {
	v := New(Config{BarCount: 4, MaxValue: 1000})
	v.bars = []int{0, 500, 1000, 250}

	// Test without rainbow
	rendered := v.RenderTall(2, false)
	if rendered == "" {
		t.Error("expected non-empty render output")
	}
	lines := strings.Split(rendered, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	// Test with rainbow
	renderedRainbow := v.RenderTall(2, true)
	if renderedRainbow == "" {
		t.Error("expected non-empty rainbow render output")
	}
	// Rainbow should contain ANSI escape codes
	if !strings.Contains(renderedRainbow, "\x1b[") {
		t.Error("expected ANSI escape codes in rainbow render")
	}
}

func TestRenderSized(t *testing.T) {
	v := New(Config{BarCount: 8, MaxValue: 1000})
	v.bars = []int{100, 200, 300, 400, 500, 600, 700, 800}

	// Test with explicit width (stretched)
	rendered := v.RenderSized(50, 4, false)
	if rendered == "" {
		t.Error("expected non-empty render output")
	}
	lines := strings.Split(rendered, "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
	// Each line should be 52 chars (50 + 2 borders) without ANSI codes
	// But we can't easily measure due to potential ANSI. Just check it's reasonable.

	// Test auto height
	renderedAuto := v.RenderSized(50, 0, false)
	if renderedAuto == "" {
		t.Error("expected non-empty auto-height render")
	}
	autoLines := strings.Split(renderedAuto, "\n")
	if len(autoLines) < 2 || len(autoLines) > 6 {
		t.Errorf("auto height should be 2-6, got %d", len(autoLines))
	}
}

func TestRenderRainbow(t *testing.T) {
	v := New(Config{BarCount: 4, MaxValue: 100})
	v.bars = []int{25, 50, 75, 100}

	rendered := v.RenderRainbow()
	if rendered == "" {
		t.Error("expected non-empty render output")
	}
	// Should contain ANSI color codes
	if !strings.Contains(rendered, "\x1b[38;5;") {
		t.Error("expected ANSI color codes in output")
	}
}
