package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/consensus"
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

// renderDetail renders the right panel (70%) with a scrollable log viewport and optional summary.
func renderDetail(
	vp viewport.Model,
	width, height int,
	role agent.Role,
	logCount int,
	isRunning bool,
	spinnerFrame int,
	bb blackboard.Blackboard,
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

	// ── Summary Section (Role Specific) ──────────────────
	summary := renderRoleSummary(role, bb, innerW)

	// ── Viewport body ────────────────────────────────────
	// Adjust viewport height based on summary height
	vpH := height - 6 // default (title + divider + border)
	if summary != "" {
		vpH -= lipgloss.Height(summary)
	}
	if vpH < 5 {
		vpH = 5
	}
	vp.Height = vpH
	body := vp.View()

	var content string
	if summary != "" {
		content = titleBar + "\n" + divider + "\n" + summary + "\n" + body
	} else {
		content = titleBar + "\n" + divider + "\n" + body
	}

	return styles.PanelBorder.
		Width(width - 2).
		Height(height - 2).
		Padding(0, 1).
		Render(content)
}

func renderRoleSummary(role agent.Role, bb blackboard.Blackboard, width int) string {
	switch role {
	case agent.RoleValidator:
		report, _ := bb.Get(consensus.KeyValidatorReport)
		semantic, _ := bb.Get(consensus.KeySemanticAgreement)
		sast, _ := bb.Get(consensus.KeySASTPassRate)

		if report == nil {
			return ""
		}

		scoreStr := fmt.Sprintf("Semantic: %.2f | SAST: %.2f", semantic, sast)
		badge := styles.ScoreBadge.Render("VALIDATION REPORT")
		scoreLine := styles.Bold.Foreground(lipgloss.Color(styles.ColSuccess)).Render(scoreStr)
		
		desc := styles.Text.Width(width - 4).Render(fmt.Sprintf("%v", report))

		return styles.SummaryBox.Width(width).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.JoinHorizontal(lipgloss.Center, badge, " ", scoreLine),
				"",
				desc,
			),
		)

	case agent.RoleArchitect:
		planSummary, _ := bb.Get(consensus.KeyArchitectPlanSummary)
		if planSummary == nil {
			return ""
		}
		badge := styles.ScoreBadge.Background(lipgloss.Color(styles.ColAccent)).Render("ARCHITECTURE PLAN")
		desc := styles.Text.Width(width - 4).Render(fmt.Sprintf("%v", planSummary))
		
		return styles.SummaryBox.Width(width).Render(
			lipgloss.JoinVertical(lipgloss.Left, badge, "", desc),
		)

	case agent.RoleNavigator:
		summary, _ := bb.Get("codebase_summary")
		if summary == nil {
			return ""
		}
		badge := styles.ScoreBadge.Background(lipgloss.Color(styles.ColSky)).Render("CODEBASE INTELLIGENCE")
		// Truncate navigator summary if too long for a summary box
		summaryStr := fmt.Sprintf("%v", summary)
		if len(summaryStr) > 200 {
			summaryStr = summaryStr[:197] + "..."
		}
		desc := styles.Text.Width(width - 4).Render(summaryStr)

		return styles.SummaryBox.Width(width).Render(
			lipgloss.JoinVertical(lipgloss.Left, badge, "", desc),
		)
	}

	return ""
}
