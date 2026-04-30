// Package styles centralises all lipgloss style definitions for the TUI.
// Consuming packages use these styles directly, ensuring a consistent theme
// without ad-hoc colour literals scattered throughout components.
package styles

import "github.com/charmbracelet/lipgloss"

// ── Colour palette ────────────────────────────────────────────────────────────
// Inspired by GitHub Dark + a purple/indigo accent that feels distinctive
// for an AI-powered developer tool.
const (
	ColPrimary  = "#7C3AED" // indigo-600  — primary accent
	ColAccent   = "#A78BFA" // violet-400  — secondary accent
	ColSuccess  = "#3FB950" // green-400   — done / pass
	ColWarning  = "#D29922" // amber-500   — warn / below threshold
	ColError    = "#F85149" // red-400     — error / failed
	ColText     = "#E6EDF3" // near-white  — body text
	ColMuted    = "#8B949E" // gray-400    — secondary text
	ColDim      = "#6E7681" // gray-500    — tertiary / disabled
	ColBorder   = "#30363D" // gray-700    — panel borders
	ColBgDark   = "#0D1117" // github-dark — header / footer background
	ColSelected = "#DDD6FE" // violet-200  — active / selected row
	ColSky      = "#58A6FF" // blue-400    — running state
	ColRunning  = "#38BDF8" // sky-400     — spinner
)

// ── Base utility styles ───────────────────────────────────────────────────────

var (
	Bold = lipgloss.NewStyle().Bold(true)
	Dim  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColDim))
	Text = lipgloss.NewStyle().Foreground(lipgloss.Color(ColText))
	Muted = lipgloss.NewStyle().Foreground(lipgloss.Color(ColMuted))

	// ── Panel borders
	PanelBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColBorder))

	ActiveBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColPrimary))

	// ── Header / footer
	HeaderStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColBgDark)).
			Foreground(lipgloss.Color(ColText)).
			Padding(0, 2)

	FooterStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColBgDark)).
			Foreground(lipgloss.Color(ColMuted)).
			Padding(0, 2)

	AppName = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColAccent)).
			Bold(true)

	KeyHint = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColAccent)).
			Bold(true)

	// ── Agent status indicators
	RunningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColRunning))
	DoneStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(ColSuccess))
	FailedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColError))
	EscalatedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColWarning))
	IdleStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(ColDim))
	SelectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColSelected)).Bold(true)

	// ── Log level labels
	LevelInfo  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColText))
	LevelWarn  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColWarning)).Bold(true)
	LevelError = lipgloss.NewStyle().Foreground(lipgloss.Color(ColError)).Bold(true)
	LevelDebug = lipgloss.NewStyle().Foreground(lipgloss.Color(ColDim))

	// ── Section headings inside panels
	SectionTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColDim)).
			Bold(true)

	// ── Input screen
	InputPrompt = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColAccent)).
			Bold(true)

	InputBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColPrimary)).
			Padding(0, 1)

	// ── Agent-specific summary styles
	SummaryBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColDim)).
			Padding(0, 1).
			MarginBottom(1)

	ScoreBadge = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Background(lipgloss.Color(ColPrimary)).
			Foreground(lipgloss.Color(ColText))

	ReportTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColAccent)).
			Bold(true).
			Underline(true)
)
