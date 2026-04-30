// Package mock provides Phase-0 mock implementations of the Agent interface.
// Each mock replays a fixed sequence of realistic log lines with randomised
// delays, driving the Bubble Tea TUI without any real LLM calls.
package mock

import (
	"context"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
)

// logPlan describes a single planned log line with its minimum emission delay.
type logPlan struct {
	delay   time.Duration
	level   string
	message string
}

// base is the shared implementation embedded by all mock agents.
// It drives the log-emission chain via: Start() → Next() → tea.Tick → LogMsg → Next() → …
//
// All methods are safe to call from the Bubble Tea main goroutine.
// The tea.Tick callbacks only construct and return new messages; they never
// mutate shared state, so no mutex is required.
type base struct {
	role   agent.Role
	status agent.Status
	plans  []logPlan
	cursor int
	// onDone is called once when all plans are exhausted, just before StatusDone
	// is emitted. Use it to write final signals to the Blackboard.
	onDone func()
}

// ── agent.Agent interface ────────────────────────────────────────────────────

func (b *base) Role() agent.Role     { return b.role }
func (b *base) Status() agent.Status { return b.status }

// Logs returns nil — in Phase 0 the TUI owns log storage in its model.
func (b *base) Logs() []agent.LogEntry { return nil }

// Reset restores the agent to its initial idle state for a fresh session.
func (b *base) Reset() {
	b.status = agent.StatusIdle
	b.cursor = 0
}

// Start marks the agent as running and immediately begins the log chain.
func (b *base) Start(_ context.Context, _, _ string) tea.Cmd {
	b.status = agent.StatusRunning
	return tea.Batch(
		func() tea.Msg { return agent.StatusMsg{Role: b.role, Status: agent.StatusRunning} },
		b.Next(),
	)
}

// Next returns the tea.Cmd for the next step in the agent's sequence.
// When all plans are exhausted it emits a StatusDone message.
func (b *base) Next() tea.Cmd {
	if b.cursor >= len(b.plans) {
		// Fire onDone hook once before transitioning to StatusDone
		if b.onDone != nil {
			b.onDone()
			b.onDone = nil // prevent double-call on Reset/re-use
		}
		b.status = agent.StatusDone
		return func() tea.Msg {
			return agent.StatusMsg{Role: b.role, Status: agent.StatusDone}
		}
	}

	plan := b.plans[b.cursor]
	b.cursor++

	// Natural jitter keeps multiple agents from firing simultaneously.
	jitter := time.Duration(rand.Intn(300)) * time.Millisecond
	delay := plan.delay + jitter

	role := b.role // capture for closure
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return agent.LogMsg{
			Role: role,
			Entry: agent.LogEntry{
				Timestamp: t,
				Level:     plan.level,
				Message:   plan.message,
			},
		}
	})
}
