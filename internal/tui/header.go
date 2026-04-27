package tui

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/tui/styles"
)

const appVersion = "v0.1.0"

// renderHeader renders the top bar showing app identity and current context.
func renderHeader(width int, workspace string, elapsed time.Duration, inputMode bool) string {
	left := styles.AppName.Render("⚡ consensus-agent") + "  " +
		styles.Dim.Render(appVersion)

	var right string
	if inputMode {
		right = styles.Muted.Render("Enter your task below")
	} else {
		// Show workspace basename (or full path if short)
		label := workspace
		if len(label) > 35 {
			label = "…/" + filepath.Base(workspace)
		}
		right = styles.Dim.Render("📁 "+label) + "  " +
			styles.Muted.Render("⏱ "+formatDuration(elapsed))
	}

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	// Fill gap between left and right
	gap := width - leftW - rightW - 4 // 4 = horizontal padding (2 each side)
	if gap < 1 {
		gap = 1
	}

	content := left + lipgloss.NewStyle().Width(gap).Render("") + right

	return styles.HeaderStyle.Width(width).Render(content)
}

// formatDuration renders an elapsed duration as MM:SS or HH:MM:SS.
func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
