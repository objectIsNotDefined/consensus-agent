package agent

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// Registry owns and provides access to all Agent instances.
type Registry struct {
	agents []Agent
	byRole map[Role]Agent
}

// NewRegistry creates a Registry from the provided agents.
// Agents are stored in the order given; use AllRoles order for consistency.
func NewRegistry(agents []Agent) *Registry {
	r := &Registry{
		agents: make([]Agent, len(agents)),
		byRole: make(map[Role]Agent, len(agents)),
	}
	copy(r.agents, agents)
	for _, a := range agents {
		r.byRole[a.Role()] = a
	}
	return r
}

// All returns the agents in their registered order.
func (r *Registry) All() []Agent {
	return r.agents
}

// GetByRole returns the agent for the given role, or nil if not registered.
func (r *Registry) GetByRole(role Role) Agent {
	return r.byRole[role]
}

// StartAll starts every agent concurrently and returns their initial commands
// to be batched together by the TUI runtime.
func (r *Registry) StartAll(ctx context.Context, task, workspace string) []tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(r.agents))
	for _, a := range r.agents {
		cmds = append(cmds, a.Start(ctx, task, workspace))
	}
	return cmds
}

// ResetAll resets every agent to its initial idle state.
func (r *Registry) ResetAll() {
	for _, a := range r.agents {
		a.Reset()
	}
}

// AllDone returns true when every agent has reached StatusDone or StatusFailed.
func (r *Registry) AllDone() bool {
	for _, a := range r.agents {
		s := a.Status()
		if s != StatusDone && s != StatusFailed {
			return false
		}
	}
	return true
}
