package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/tui/styles"
)

// renderLog renders a single log entry: timestamp + coloured level badge + message.
func renderLog(entry agent.LogEntry) string {
	ts := styles.Dim.Render(entry.Timestamp.Format("15:04:05"))

	var levelStyle lipgloss.Style
	switch entry.Level {
	case "WARN":
		levelStyle = styles.LevelWarn
	case "ERROR":
		levelStyle = styles.LevelError
	case "DEBUG":
		levelStyle = styles.LevelDebug
	default:
		levelStyle = styles.LevelInfo
	}
	level := levelStyle.Render(fmt.Sprintf("[%-5s]", entry.Level))
	msg := styles.Text.Render(entry.Message)

	return fmt.Sprintf("%s %s %s", ts, level, msg)
}

// buildDetailContent builds the full scrollable log content string for the viewport.
func buildDetailContent(logs []agent.LogEntry) string {
	if len(logs) == 0 {
		return styles.Muted.Render("  Waiting for agent to start...")
	}
	var sb strings.Builder
	for _, e := range logs {
		sb.WriteString("  ")
		sb.WriteString(renderLog(e))
		sb.WriteString("\n")
	}
	return sb.String()
}

// renderDetail renders the right panel (70%) with a scrollable log viewport.
// Width and height are total outer dimensions including the panel border.
func renderDetail(
	vp viewport.Model,
	width, height int,
	role agent.Role,
	logCount int,
	isRunning bool,
	spinnerFrame int,
) string {
	innerW := width - 4 // border(2) + padding(2)

	// ── Title bar ────────────────────────────────────────
	var statusTag string
	if isRunning {
		statusTag = styles.RunningStyle.
			Render(" " + spinnerFrames[spinnerFrame%len(spinnerFrames)])
	}
	titleLeft := styles.AppName.Render("▶ " + string(role)) + statusTag
	titleRight := styles.Muted.Render(fmt.Sprintf("%d lines", logCount))

	titlePad := innerW - lipgloss.Width(titleLeft) - lipgloss.Width(titleRight)
	if titlePad < 1 {
		titlePad = 1
	}
	titleBar := titleLeft + strings.Repeat(" ", titlePad) + titleRight

	// ── Divider ──────────────────────────────────────────
	divider := styles.Dim.Render(strings.Repeat("─", innerW))

	// ── Viewport body ────────────────────────────────────
	body := vp.View()

	content := titleBar + "\n" + divider + "\n" + body

	return styles.PanelBorder.
		Width(width - 2).
		Height(height - 2).
		Padding(0, 1).
		Render(content)
}
