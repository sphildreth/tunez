package artwork

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCacheGetSet(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewCache(dir, 30)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	// Test miss
	if _, ok := cache.Get("ref1", 20); ok {
		t.Error("expected cache miss")
	}

	// Test set and get
	if err := cache.Set("ref1", 20, "test ansi art"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	ansi, ok := cache.Get("ref1", 20)
	if !ok {
		t.Error("expected cache hit")
	}
	if ansi != "test ansi art" {
		t.Errorf("expected 'test ansi art', got %q", ansi)
	}

	// Different width should be a miss
	if _, ok := cache.Get("ref1", 10); ok {
		t.Error("expected cache miss for different width")
	}
}

func TestCacheClear(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewCache(dir, 30)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	cache.Set("ref1", 20, "art1")
	cache.Set("ref2", 20, "art2")

	if err := cache.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	if _, ok := cache.Get("ref1", 20); ok {
		t.Error("expected cache miss after clear")
	}
}

func TestCacheSize(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewCache(dir, 30)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	cache.Set("ref1", 20, "short")
	cache.Set("ref2", 20, strings.Repeat("x", 1000))

	size, err := cache.Size()
	if err != nil {
		t.Fatalf("Size: %v", err)
	}
	if size < 1000 {
		t.Errorf("expected size >= 1000, got %d", size)
	}
}

func TestConvertToANSI(t *testing.T) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// Fill with red
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	// Encode to PNG
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("encode png: %v", err)
	}
	f.Close()

	// Read and convert
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	ansi, err := ConvertToANSI(context.Background(), data, 5, 5)
	if err != nil {
		t.Fatalf("ConvertToANSI: %v", err)
	}

	// Should contain ANSI escape codes
	if !strings.Contains(ansi, "\x1b[") {
		t.Error("expected ANSI escape codes in output")
	}
}

func TestPlaceholder(t *testing.T) {
	ph := Placeholder(20, 10)
	if !strings.Contains(ph, "♪") {
		t.Error("expected music note in placeholder")
	}
	if !strings.Contains(ph, "┌") {
		t.Error("expected border in placeholder")
	}
}

func TestRgbTo256(t *testing.T) {
	tests := []struct {
		r, g, b uint8
		want    int
	}{
		{0, 0, 0, 16},        // black -> grayscale
		{255, 255, 255, 231}, // white -> grayscale
		{255, 0, 0, 196},     // red
		{0, 255, 0, 46},      // green
		{0, 0, 255, 21},      // blue
		{128, 128, 128, 244}, // gray -> grayscale ramp
	}

	for _, tt := range tests {
		got := rgbTo256(tt.r, tt.g, tt.b)
		// Allow some tolerance for color approximation
		if got < 16 || got > 255 {
			t.Errorf("rgbTo256(%d,%d,%d) = %d, out of range", tt.r, tt.g, tt.b, got)
		}
	}
}

func TestDefaultArtwork(t *testing.T) {
	// Should return ANSI art from embedded PNG
	art := DefaultArtwork(20, 10)
	if art == "" {
		t.Error("expected non-empty default artwork")
	}

	// Should contain ANSI escape codes
	if !strings.Contains(art, "\x1b[") {
		t.Error("expected ANSI escape codes in default artwork")
	}

	// Should be cached (call again and verify same result)
	art2 := DefaultArtwork(20, 10)
	if art != art2 {
		t.Error("expected cached result to match")
	}
}
