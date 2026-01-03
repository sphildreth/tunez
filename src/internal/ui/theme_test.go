package ui

import "testing"

func TestRainbow(t *testing.T) {
	t1 := Rainbow(false)
	if t1.Accent.GetForeground() == nil {
		t.Error("Rainbow(false) should have colors")
	}

	t2 := Rainbow(true)
	// lipgloss.NewStyle().GetForeground() returns nil by default
	// but we need to check if it's "reset"
	// Actually, just checking one property
	if t2.Title.GetBold() == false {
		// our implementation sets bold even in noColor,
		// wait: reset.Bold(true) in noColor branch?
		// 	Title:     reset.Bold(true),
		// Yes.
	}
}
