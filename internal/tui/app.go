// Package tui implements the full-screen Bubble Tea TUI for consensus-agent.
//
// Layout (run mode):
//
//	┌─ Header ─────────────────────────────────────────────────────────┐
//	├─ Overview (30%) ─────────┬─ Agent Detail (70%) ─────────────────┤
//	│ AGENTS                   │ ▶ Architect                          │
//	│ ▶ ✓ Navigator 🧭        │ 12:01:01 [INFO] Decomposing task...  │
//	│   ⠸ Architect 🏗        │ 12:01:02 [WARN] Score below 0.85    │
//	│   ⠸ Executor  ⚙        │ …                                    │
//	│   ⠸ Validator 🛡        │                                      │
//	│                          │                                      │
//	│ CONSENSUS SCORE          │                                      │
//	│ ████████░░ 0.82 / 0.85  │                                      │
//	│ TOKEN BUDGET             │                                      │
//	│ ██████░░░░ 6.2k / 10k   │                                      │
//	├─ Footer ─────────────────┴──────────────────────────────────────┤
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/consensus"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/dag"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/tui/styles"
	"github.com/objectisnotdefined/consensus-agent/ca/pkg/config"
)

// tickMsg drives the spinner animation and elapsed-time counter.
type tickMsg time.Time

func scheduleTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ── Model ────────────────────────────────────────────────────────────────────

// Model is the root Bubble Tea model that owns all sub-components and drives
// the consensus-agent TUI lifecycle.
type Model struct {
	// ── Dependencies
	registry  *agent.Registry
	dagExec   *dag.Executor
	evaluator *consensus.Evaluator
	workspace string
	cfg       *config.Config
	bb        blackboard.Blackboard
	sessionID string

	// ── Terminal dimensions
	width  int
	height int

	// ── Application state
	inputMode    bool // true → initial task input screen
	followUpMode bool // true → show input box at the bottom of the dashboard
	allDone      bool // true when consensus is reached or escalated
	escalated    bool // true when max debate rounds exhausted without consensus
	debateRound  int  // current debate round (0 = not yet evaluated)

	// ── Current task (stored so follow-up rounds can re-use DAG)
	task string

	// ── Task input
	input textinput.Model

	// ── Per-agent state (owned by TUI, not agents)
	statuses map[agent.Role]agent.Status
	logs     map[agent.Role][]agent.LogEntry

	// ── Active agent panel
	activeRole agent.Role

	// ── Log viewport (right panel)
	viewport      viewport.Model
	viewportReady bool

	// ── Metrics / animation
	spinnerFrame   int
	startTime      time.Time
	elapsed        time.Duration
	tokenUsed      int
	consensusScore float64
	turns          int

	// ── Context (for future agent cancellation)
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates the root model. Call tea.NewProgram(model, tea.WithAltScreen()) to run.
func New(registry *agent.Registry, dagExec *dag.Executor, evaluator *consensus.Evaluator, workspace string, cfg *config.Config, bb blackboard.Blackboard, sessionID string) *Model {
	ti := textinput.New()
	ti.Placeholder = "e.g. Add JWT auth middleware to the HTTP router"
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 60

	statuses := make(map[agent.Role]agent.Status, len(agent.AllRoles))
	logs := make(map[agent.Role][]agent.LogEntry, len(agent.AllRoles))
	for _, r := range agent.AllRoles {
		statuses[r] = agent.StatusIdle
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Model{
		registry:   registry,
		dagExec:    dagExec,
		evaluator:  evaluator,
		workspace:  workspace,
		cfg:        cfg,
		bb:         bb,
		sessionID:  sessionID,
		inputMode:  true,
		input:      ti,
		statuses:   statuses,
		logs:       logs,
		activeRole: agent.RoleNavigator,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// ── tea.Model implementation ─────────────────────────────────────────────────

// Init runs once at startup.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, scheduleTick())
}

// Update handles all incoming messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// ── Window size ──────────────────────────────────────
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.rebuildViewport()

	// ── Animation tick ───────────────────────────────────
	case tickMsg:
		m.spinnerFrame++
		if !m.inputMode {
			m.elapsed = time.Since(m.startTime)
		}
		cmds = append(cmds, scheduleTick())

	// ── Agent log line ───────────────────────────────────
	case agent.LogMsg:
		m.logs[msg.Role] = append(m.logs[msg.Role], msg.Entry)
		m.tokenUsed += 120 + len(msg.Entry.Message)*2 // simulate token growth
		if msg.Role == m.activeRole {
			m.syncViewport()
		}
		// Schedule next log in the chain
		if a := m.registry.GetByRole(msg.Role); a != nil {
			cmds = append(cmds, a.Next())
		}

	// ── Agent status change ──────────────────────────────
	case agent.StatusMsg:
		m.statuses[msg.Role] = msg.Status
		if msg.Status == agent.StatusDone || msg.Status == agent.StatusFailed {
			if msg.Status == agent.StatusDone {
				// Ask the DAG executor which agents are now unblocked
				nextRoles := m.dagExec.MarkDone(msg.Role)
				for _, role := range nextRoles {
					if cmd := m.registry.StartRole(m.ctx, role, m.task, m.workspace); cmd != nil {
						cmds = append(cmds, cmd)
					}
				}
			}
			if m.registry.AllDone() {
				// All agents done — run the real consensus evaluation
				report := m.evaluator.Evaluate()
				m.consensusScore = report.FinalScore
				m.debateRound = report.Round

				if report.Passed {
					// ✅ Consensus reached
					m.allDone = true
				} else if m.evaluator.ShouldEscalate(report) {
					// 🚨 Max rounds exhausted — escalate to Human-in-the-Loop
					m.allDone = true
					m.escalated = true
				} else {
					// 🔄 Debate round: reset Executor + Validator and re-run
					debateRoles := []agent.Role{agent.RoleExecutor, agent.RoleValidator}
					m.registry.ResetRoles(debateRoles)
					m.dagExec.ResetRoles(debateRoles)
					for _, r := range debateRoles {
						m.statuses[r] = agent.StatusIdle
						m.logs[r] = nil
					}
					for _, role := range m.dagExec.Ready() {
						if cmd := m.registry.StartRole(m.ctx, role, m.task, m.workspace); cmd != nil {
							cmds = append(cmds, cmd)
						}
					}
				}
			}
		}

	// ── Keyboard ─────────────────────────────────────────
	case tea.KeyMsg:
		if m.inputMode {
			cmds = append(cmds, m.handleInputKey(msg)...)
		} else {
			cmds = append(cmds, m.handleRunKey(msg)...)
		}
	}

	// Delegate to child models
	if m.inputMode || m.followUpMode {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}
	if !m.inputMode && m.viewportReady {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the full TUI frame.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	// Guard for tiny terminals
	if m.width < 60 || m.height < 15 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColWarning)).
			Padding(1, 2).
			Render("⚠  Terminal too small. Please resize to at least 60×15.")
	}
	if m.inputMode {
		return m.renderInputScreen()
	}
	return m.renderRunScreen()
}

// ── Key handlers ─────────────────────────────────────────────────────────────

func (m *Model) handleInputKey(msg tea.KeyMsg) []tea.Cmd {
	switch msg.Type {
	case tea.KeyEnter:
		task := strings.TrimSpace(m.input.Value())
		if task == "" {
			return nil
		}
		// Switch to run mode
		m.task = task
		m.inputMode = false
		m.input.Blur()
		m.startTime = time.Now()
		m.rebuildViewport()
		_, _ = m.bb.CreateTurn(m.sessionID, task, m.turns)
		// Start only root nodes (Navigator) — DAG will unlock the rest
		var cmds []tea.Cmd
		for _, role := range m.dagExec.Ready() {
			if cmd := m.registry.StartRole(m.ctx, role, task, m.workspace); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return cmds

	case tea.KeyCtrlC, tea.KeyEsc:
		m.cancel()
		return []tea.Cmd{tea.Quit}
	}
	return nil
}

func (m *Model) handleRunKey(msg tea.KeyMsg) []tea.Cmd {
	if m.followUpMode {
		switch msg.Type {
		case tea.KeyEnter:
			task := strings.TrimSpace(m.input.Value())
			if task == "" {
				return nil
			}
			m.task = task
			m.followUpMode = false
			m.turns++
			m.allDone = false
			m.consensusScore = 0
			m.elapsed = 0
			m.startTime = time.Now()
			for _, r := range agent.AllRoles {
				m.statuses[r] = agent.StatusIdle
				m.logs[r] = nil
			}
			m.registry.ResetAll()
			m.dagExec.Reset()
			m.input.Blur()
			m.rebuildViewport()
			_, _ = m.bb.CreateTurn(m.sessionID, task, m.turns)
			// Start only root nodes — DAG drives the rest
			var cmds []tea.Cmd
			for _, role := range m.dagExec.Ready() {
				if cmd := m.registry.StartRole(m.ctx, role, task, m.workspace); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
			return cmds
		case tea.KeyEsc:
			m.followUpMode = false
			m.input.Blur()
			return nil
		}
		// In follow-up mode, we only handle Enter and Esc.
		// Other keys are passed to the textinput component in the main Update loop.
		return nil
	}

	switch msg.String() {
	case "q", "Q", "ctrl+c":
		m.cancel()
		return []tea.Cmd{tea.Quit}
	case "tab":
		m.cycleAgent()
		m.syncViewport()
	case "r", "R":
		m.resetSession()
	case "enter":
		if m.allDone {
			m.followUpMode = true
			m.input.Reset()
			m.input.Placeholder = fmt.Sprintf("Turn %d: What's next?", m.turns+2)
			m.input.Focus()
			return []tea.Cmd{textinput.Blink}
		}
	}
	return nil
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (m *Model) cycleAgent() {
	for i, r := range agent.AllRoles {
		if r == m.activeRole {
			m.activeRole = agent.AllRoles[(i+1)%len(agent.AllRoles)]
			return
		}
	}
}

// rebuildViewport creates or resizes the viewport to fit the current terminal.
func (m *Model) rebuildViewport() {
	if m.width == 0 || m.height == 0 {
		return
	}
	// Reserve 1 line each for header and footer.
	panelH := m.height - 2
	leftW := m.width * 3 / 10
	rightW := m.width - leftW

	// Inner viewport: border(2) + padding(2) + title(1) + divider(1) = 6
	vpW := rightW - 6
	vpH := panelH - 6
	if vpW < 10 {
		vpW = 10
	}
	if vpH < 3 {
		vpH = 3
	}

	m.viewport = viewport.New(vpW, vpH)
	m.viewport.SetContent(buildDetailContent(m.logs[m.activeRole]))
	m.viewport.GotoBottom()
	m.viewportReady = true
}

// syncViewport refreshes the viewport content for the currently active agent.
func (m *Model) syncViewport() {
	if !m.viewportReady {
		return
	}
	m.viewport.SetContent(buildDetailContent(m.logs[m.activeRole]))
	m.viewport.GotoBottom()
}

// resetSession resets the TUI and all agents for a new task.
func (m *Model) resetSession() {
	m.turns = 0
	m.registry.ResetAll()
	m.dagExec.Reset()
	m.evaluator.Reset()
	m.inputMode = true
	m.allDone = false
	m.escalated = false
	m.debateRound = 0
	m.tokenUsed = 0
	m.consensusScore = 0
	m.elapsed = 0
	for _, r := range agent.AllRoles {
		m.statuses[r] = agent.StatusIdle
		m.logs[r] = nil
	}
	m.activeRole = agent.RoleNavigator
	m.input.Reset()
	m.input.Focus()
	m.rebuildViewport()
}

// ── Screen renderers ─────────────────────────────────────────────────────────

// renderInputScreen renders the centered task-input screen.
func (m *Model) renderInputScreen() string {
	header := renderHeader(m.width, m.workspace, 0, m.turns, true)
	footer := renderFooter(m.width, true)
	bodyH := m.height - 2 // remove header and footer lines

	// Input box width: narrower than terminal, minimum 50
	boxW := m.width - 20
	if boxW > 80 {
		boxW = 80
	}
	if boxW < 50 {
		boxW = 50
	}

	title := styles.AppName.Bold(true).Render("⚡  consensus-agent")
	subtitle := styles.Muted.Render("Model Consensus Driven Development  ·  MCDD")
	wsLine := styles.Dim.Render("📁  Workspace: ") + styles.Text.Render(m.workspace)

	promptLine := "\n" + styles.InputPrompt.Render("What should the agents do today?")
	inputWidget := styles.InputBox.Width(boxW).Render(m.input.View())
	hint := "\n" + styles.Muted.Render("[Enter]") + "  " + styles.Dim.Render("to start") +
		"     " + styles.Muted.Render("[Ctrl+C]") + "  " + styles.Dim.Render("to quit")

	block := lipgloss.JoinVertical(lipgloss.Left,
		title, subtitle, "", wsLine, promptLine, "", inputWidget, hint,
	)

	body := lipgloss.Place(m.width, bodyH, lipgloss.Center, lipgloss.Center, block)

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// renderRunScreen renders the 2-panel agent monitoring screen.
func (m *Model) renderRunScreen() string {
	panelH := m.height - 2 // reserve 1 line each for header + footer
	leftW := m.width * 3 / 10
	rightW := m.width - leftW

	// Build log-count map
	logCounts := make(map[agent.Role]int, len(agent.AllRoles))
	for _, r := range agent.AllRoles {
		logCounts[r] = len(m.logs[r])
	}

	header := renderHeader(m.width, m.workspace, m.elapsed, m.turns, false)

	// Footer logic
	var footer string
	if m.followUpMode {
		prompt := styles.InputPrompt.Render(fmt.Sprintf(" Turn %d > ", m.turns+2))
		footer = lipgloss.NewStyle().
			Width(m.width).
			Background(lipgloss.Color(styles.ColBgDark)).
			Padding(0, 1).
			Render(lipgloss.JoinHorizontal(lipgloss.Center, prompt, m.input.View()))
	} else if m.allDone {
		footer = renderConsensusBanner(m.width, m.consensusScore, m.elapsed, m.debateRound, m.escalated)
	} else {
		footer = renderFooter(m.width, false)
	}

	left := renderOverview(
		leftW, panelH,
		m.registry.All(),
		m.statuses,
		m.activeRole,
		logCounts,
		m.spinnerFrame,
		m.consensusScore,
		m.cfg.Consensus.Threshold,
		m.tokenUsed,
		m.cfg.Cost.TokenBudget,
	)

	var right string
	if m.viewportReady {
		isRunning := m.statuses[m.activeRole] == agent.StatusRunning
		right = renderDetail(m.viewport, rightW, panelH, m.activeRole,
			logCounts[m.activeRole], isRunning, m.spinnerFrame)
	} else {
		right = styles.PanelBorder.Width(rightW - 2).Height(panelH - 2).Render("")
	}

	panels := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return lipgloss.JoinVertical(lipgloss.Left, header, panels, footer)
}

// renderConsensusBanner renders the banner shown when evaluation completes.
// It adapts based on whether consensus was reached, escalated, or pending debate.
func renderConsensusBanner(width int, score float64, elapsed time.Duration, debateRound int, escalated bool) string {
	scoreFmt := lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColSuccess)).Render(formatScore(score))
	roundFmt := fmt.Sprintf("Round %d", debateRound)

	var title, detail, bg string
	if escalated {
		title = styles.AppName.Bold(true).Foreground(lipgloss.Color(styles.ColWarning)).Render("⚠  Escalated to Human Review")
		detail = "  ·  Score: " + scoreFmt + "  ·  " + roundFmt + "  ·  [Enter] Provide guidance  [R] Reset  [Q] Quit"
		bg = "#2D1A00"
	} else {
		title = styles.DoneStyle.Bold(true).Render("✅  Consensus Reached")
		detail = "  ·  Score: " + scoreFmt + "  ·  Time: " + formatDuration(elapsed) + "  ·  " + roundFmt + "  ·  [Enter] Next turn  [R] Reset  [Q] Quit"
		bg = "#0D2818"
	}

	msg := title + styles.Muted.Render(detail)
	return lipgloss.NewStyle().
		Width(width).
		Background(lipgloss.Color(bg)).
		Padding(0, 2).
		Render(msg)
}

func formatScore(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
