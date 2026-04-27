// Package agent defines the core types and interfaces for all agents in the
// MCDD (Model Consensus Driven Development) pipeline.
package agent

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Role identifies an agent's specialization in the MCDD pipeline.
type Role string

const (
	RoleNavigator Role = "Navigator"
	RoleArchitect Role = "Architect"
	RoleExecutor  Role = "Executor"
	RoleValidator Role = "Validator"
)

// AllRoles is the canonical display order of agents.
var AllRoles = []Role{RoleNavigator, RoleArchitect, RoleExecutor, RoleValidator}

// RoleIcon returns the display icon for a role.
func RoleIcon(r Role) string {
	switch r {
	case RoleNavigator:
		return "🧭"
	case RoleArchitect:
		return "🏗"
	case RoleExecutor:
		return "⚙"
	case RoleValidator:
		return "🛡"
	default:
		return "◉"
	}
}

// Status represents an agent's lifecycle state.
type Status int

const (
	StatusIdle Status = iota
	StatusRunning
	StatusDone
	StatusFailed
	StatusEscalated
)

func (s Status) String() string {
	return [...]string{"Idle", "Running", "Done", "Failed", "Escalated"}[s]
}

// LogEntry is a single timestamped log line emitted by an agent.
type LogEntry struct {
	Timestamp time.Time
	Level     string // "INFO" | "WARN" | "ERROR" | "DEBUG"
	Message   string
}

// LogMsg is a Bubble Tea message emitted when an agent produces a new log line.
// After the TUI receives this, it calls agent.Next() to schedule the next log.
type LogMsg struct {
	Role  Role
	Entry LogEntry
}

// StatusMsg is a Bubble Tea message emitted when an agent changes state.
type StatusMsg struct {
	Role   Role
	Status Status
}

// Agent is the minimal interface all pipeline agents must implement.
//
// Phase 0: implemented by mock agents (no LLM calls).
// Phase 1: implemented by real LLM-backed agents (Navigator, Architect, etc.).
type Agent interface {
	Role() Role
	Status() Status
	Logs() []LogEntry

	// Start kicks off the agent's work for the given task and workspace.
	// Returns the first tea.Cmd in the log-emission chain.
	Start(ctx context.Context, task, workspace string) tea.Cmd

	// Next returns the tea.Cmd to emit the next log entry.
	// The TUI calls this each time it receives a LogMsg from this agent.
	// Returns a StatusMsg{Done} cmd when the agent has no more steps.
	Next() tea.Cmd

	// Reset clears all state so the agent can be reused for a new session.
	Reset()
}
