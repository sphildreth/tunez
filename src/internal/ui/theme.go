package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Accent    lipgloss.Style
	Dim       lipgloss.Style
	Text      lipgloss.Style
	Title     lipgloss.Style
	Error     lipgloss.Style
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Border    lipgloss.Style
	Highlight lipgloss.Style
}

func Rainbow(noColor bool) Theme {
	reset := lipgloss.NewStyle()
	if noColor {
		return Theme{
			Accent:    reset,
			Dim:       reset,
			Text:      reset,
			Title:     reset.Bold(true),
			Error:     reset,
			Success:   reset,
			Warning:   reset,
			Border:    reset,
			Highlight: reset.Reverse(true),
		}
	}
	return Theme{
		Accent:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6FF7")),
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6F93")),
		Text:      lipgloss.NewStyle().Foreground(lipgloss.Color("#E6E6FA")),
		Title:     lipgloss.NewStyle().Foreground(lipgloss.Color("#8EEBFF")).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F56")).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(lipgloss.Color("#5CFF5C")).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD166")).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(lipgloss.Color("#7C7CFF")),
		Highlight: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA7C4")).Bold(true),
	}
}
