package mock

import (
	"time"

	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
)

// Navigator simulates the Navigator role: scanning the workspace, building a
// semantic code index, and publishing a codebase snapshot to the Blackboard.
type Navigator struct {
	base
	bb blackboard.Blackboard
}

// NewNavigator returns a mock Navigator agent.
func NewNavigator(bb blackboard.Blackboard) agent.Agent {
	plans := []logPlan{
		{350 * time.Millisecond, "INFO", "Initializing workspace scan..."},
		{500 * time.Millisecond, "INFO", "Traversing directory tree recursively"},
		{600 * time.Millisecond, "INFO", "Discovered: 47 Go files · 3 config files · 12 test files"},
		{700 * time.Millisecond, "INFO", "Building semantic code index..."},
		{800 * time.Millisecond, "INFO", "Parsing AST for all source files"},
		{600 * time.Millisecond, "INFO", "Indexing function signatures and type definitions"},
		{500 * time.Millisecond, "INFO", "Indexing interface contracts and dependencies"},
		{400 * time.Millisecond, "INFO", "Context window usage: 18%  (183k / 1M tokens)"},
		{350 * time.Millisecond, "INFO", "Publishing codebase snapshot to Blackboard"},
		{300 * time.Millisecond, "INFO", "Navigation graph ready — 312 nodes, 891 edges"},
		{200 * time.Millisecond, "INFO", "✅ Navigator complete. Blackboard updated."},
	}
	return &Navigator{
		base: base{role: agent.RoleNavigator, plans: plans},
		bb:   bb,
	}
}
