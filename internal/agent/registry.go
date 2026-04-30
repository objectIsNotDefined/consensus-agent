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
//
// Deprecated: prefer StartRole for DAG-driven orchestration.
func (r *Registry) StartAll(ctx context.Context, task, workspace string) []tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(r.agents))
	for _, a := range r.agents {
		cmds = append(cmds, a.Start(ctx, task, workspace))
	}
	return cmds
}

// StartRole starts the agent for a single role and returns its initial tea.Cmd.
// Returns nil if no agent is registered for the given role.
func (r *Registry) StartRole(ctx context.Context, role Role, task, workspace string) tea.Cmd {
	a := r.byRole[role]
	if a == nil {
		return nil
	}
	return a.Start(ctx, task, workspace)
}

// ResetAll resets every agent to its initial idle state.
func (r *Registry) ResetAll() {
	for _, a := range r.agents {
		a.Reset()
	}
}

// ResetRoles resets only the agents for the specified roles.
// Used by the Debate Loop to re-run only Executor and Validator.
func (r *Registry) ResetRoles(roles []Role) {
	for _, role := range roles {
		if a := r.byRole[role]; a != nil {
			a.Reset()
		}
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
