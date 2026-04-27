package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/tui/styles"
)

// Braille-dot spinner frames — classic terminal animation.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// renderOverview renders the left panel (30%): agent statuses, consensus score
// and token budget. Width and height are the total outer dimensions including
// the panel border.
func renderOverview(
	width, height int,
	agents []agent.Agent,
	statuses map[agent.Role]agent.Status,
	activeRole agent.Role,
	logCounts map[agent.Role]int,
	spinnerFrame int,
	consensusScore float64,
	consensusThreshold float64,
	tokenUsed, tokenBudget int,
) string {
	// Inner content width (border consumes 2: one each side)
	innerW := width - 4
	if innerW < 8 {
		innerW = 8
	}

	var sb strings.Builder

	// ── AGENTS ───────────────────────────────────────────
	sb.WriteString(styles.SectionTitle.Render("AGENTS"))
	sb.WriteString("\n\n")

	for _, a := range agents {
		status := statuses[a.Role()]
		row := renderAgentRow(a.Role(), status, logCounts[a.Role()], spinnerFrame, a.Role() == activeRole)
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// ── CONSENSUS SCORE ───────────────────────────────────
	sb.WriteString(styles.SectionTitle.Render("CONSENSUS SCORE"))
	sb.WriteString("\n")

	barW := innerW - 2
	if barW < 4 {
		barW = 4
	}
	scoreBar := renderProgressBar(consensusScore, 1.0, barW, styles.ColPrimary, styles.ColBorder)
	sb.WriteString(scoreBar)
	sb.WriteString("\n")

	scoreColor := styles.ColError
	if consensusScore >= consensusThreshold {
		scoreColor = styles.ColSuccess
	} else if consensusScore >= consensusThreshold*0.75 {
		scoreColor = styles.ColWarning
	}
	sb.WriteString(
		lipgloss.NewStyle().Foreground(lipgloss.Color(scoreColor)).Bold(true).
			Render(fmt.Sprintf("%.2f", consensusScore)) +
			styles.Muted.Render(fmt.Sprintf(" / %.2f", consensusThreshold)),
	)
	sb.WriteString("\n\n")

	// ── TOKEN BUDGET ──────────────────────────────────────
	sb.WriteString(styles.SectionTitle.Render("TOKEN BUDGET"))
	sb.WriteString("\n")

	tokenRatio := 0.0
	if tokenBudget > 0 {
		tokenRatio = float64(tokenUsed) / float64(tokenBudget)
	}
	tokenColor := styles.ColAccent
	if tokenRatio > 0.8 {
		tokenColor = styles.ColWarning
	}
	tokenBar := renderProgressBar(tokenRatio, 1.0, barW, tokenColor, styles.ColBorder)
	sb.WriteString(tokenBar)
	sb.WriteString("\n")
	sb.WriteString(
		styles.Muted.Render(formatTokens(tokenUsed)) +
			styles.Dim.Render(" / "+formatTokens(tokenBudget)),
	)

	content := sb.String()

	return styles.PanelBorder.
		Width(width - 2).
		Height(height - 2).
		Padding(0, 1).
		Render(content)
}

// renderAgentRow renders a single agent status row in the overview panel.
func renderAgentRow(
	role agent.Role,
	status agent.Status,
	logCount int,
	spinnerFrame int,
	active bool,
) string {
	// Status icon / spinner
	var icon string
	switch status {
	case agent.StatusRunning:
		icon = styles.RunningStyle.Render(spinnerFrames[spinnerFrame%len(spinnerFrames)])
	case agent.StatusDone:
		icon = styles.DoneStyle.Render("✓")
	case agent.StatusFailed:
		icon = styles.FailedStyle.Render("✗")
	case agent.StatusEscalated:
		icon = styles.EscalatedStyle.Render("!")
	default:
		icon = styles.IdleStyle.Render("○")
	}

	// Role name styling
	nameStyle := styles.IdleStyle
	if active {
		nameStyle = styles.SelectedStyle
	} else if status == agent.StatusRunning {
		nameStyle = styles.Text
	} else if status == agent.StatusDone {
		nameStyle = styles.DoneStyle
	}

	roleIcon := agent.RoleIcon(role)
	name := fmt.Sprintf("%-10s", string(role))
	nameStr := nameStyle.Render(roleIcon + " " + name)

	// Log count badge
	badge := ""
	if logCount > 0 {
		badge = styles.Dim.Render(fmt.Sprintf("%d lines", logCount))
	}

	// Selection indicator
	selector := "  "
	if active {
		selector = styles.KeyHint.Render("▶ ")
	}

	return fmt.Sprintf("%s%s %s  %s", selector, icon, nameStr, badge)
}

// renderProgressBar renders an ASCII progress bar.
// value and max define the ratio; width is the total bar character width.
func renderProgressBar(value, max float64, width int, filledCol, emptyCol string) string {
	if width < 2 {
		width = 2
	}
	ratio := 0.0
	if max > 0 {
		ratio = value / max
	}
	if ratio > 1.0 {
		ratio = 1.0
	}
	filled := int(ratio * float64(width))
	empty := width - filled

	filledStr := lipgloss.NewStyle().Foreground(lipgloss.Color(filledCol)).
		Render(strings.Repeat("█", filled))
	emptyStr := lipgloss.NewStyle().Foreground(lipgloss.Color(emptyCol)).
		Render(strings.Repeat("░", empty))
	return filledStr + emptyStr
}

// formatTokens formats a token count as "1.2k" or plain integer.
func formatTokens(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%d", n)
}
