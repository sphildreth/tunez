package artwork

import (
	"context"
	"strings"
	"testing"
)

// TestKittyOutputStructure verifies that Kitty protocol output has the correct
// structure for TUI frameworks - it should have proper line counts so that
// lipgloss can measure dimensions correctly.
func TestKittyOutputStructure(t *testing.T) {
	// Use the embedded default artwork PNG
	testPNG := defaultArtworkPNG

	widthCells := 10
	heightCells := 5

	// Test Kitty output
	kittyResult, err := ConvertToKitty(context.Background(), testPNG, widthCells, heightCells)
	if err != nil {
		t.Fatalf("ConvertToKitty failed: %v", err)
	}

	// Test ANSI output for comparison
	ansiResult, err := ConvertToANSI(context.Background(), testPNG, widthCells, heightCells, QualityMedium, ScaleFit)
	if err != nil {
		t.Fatalf("ConvertToANSI failed: %v", err)
	}

	// Analyze Kitty output
	kittyLines := strings.Split(kittyResult, "\n")
	ansiLines := strings.Split(ansiResult, "\n")

	t.Logf("=== Kitty Output ===")
	t.Logf("Total bytes: %d", len(kittyResult))
	t.Logf("Line count: %d", len(kittyLines))
	t.Logf("Newline count: %d", strings.Count(kittyResult, "\n"))
	t.Logf("Has Kitty escape: %v", strings.Contains(kittyResult, "\x1b_G"))

	t.Logf("=== ANSI Output ===")
	t.Logf("Total bytes: %d", len(ansiResult))
	t.Logf("Line count: %d", len(ansiLines))
	t.Logf("Newline count: %d", strings.Count(ansiResult, "\n"))

	// The key issue: Kitty output should have the same number of lines as ANSI
	// for lipgloss to measure correctly
	t.Logf("=== Comparison ===")
	t.Logf("ANSI lines: %d, Kitty lines: %d, requested height: %d", len(ansiLines), len(kittyLines), heightCells)

	// For TUI compatibility, Kitty output MUST have the same line count as ANSI
	if len(kittyLines) != len(ansiLines) {
		t.Errorf("Line count mismatch: ANSI=%d, Kitty=%d - Kitty needs same lines for TUI layout", len(ansiLines), len(kittyLines))
	}
}

// TestKittyVsANSILineCount compares line counts at various dimensions
func TestKittyVsANSILineCount(t *testing.T) {
	testPNG := defaultArtworkPNG

	testCases := []struct {
		width, height int
	}{
		{10, 5},
		{20, 10},
		{40, 20},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			kitty, err := ConvertToKitty(context.Background(), testPNG, tc.width, tc.height)
			if err != nil {
				t.Fatalf("ConvertToKitty failed: %v", err)
			}
			ansi, err := ConvertToANSI(context.Background(), testPNG, tc.width, tc.height, QualityMedium, ScaleFit)
			if err != nil {
				t.Fatalf("ConvertToANSI failed: %v", err)
			}

			kittyLines := strings.Split(kitty, "\n")
			ansiLines := strings.Split(ansi, "\n")

			t.Logf("Dims %dx%d: ANSI=%d lines, Kitty=%d lines",
				tc.width, tc.height, len(ansiLines), len(kittyLines))

			// Kitty must match ANSI line count for TUI compatibility
			if len(kittyLines) != len(ansiLines) {
				t.Errorf("Line count mismatch: ANSI=%d, Kitty=%d (need to match for TUI)",
					len(ansiLines), len(kittyLines))
			}
		})
	}
}
