package mock

import (
	"time"

	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
)

// Executor simulates the Executor role: implementing code logic for subtasks
// assigned by the Architect inside an isolated virtual workspace sandbox.
type Executor struct {
	base
	bb blackboard.Blackboard
}

// NewExecutor returns a mock Executor agent.
func NewExecutor(bb blackboard.Blackboard) agent.Agent {
	plans := []logPlan{
		{400 * time.Millisecond, "INFO", "Awaiting subtask assignment from Architect..."},
		{600 * time.Millisecond, "INFO", "SubTask [A] received: Define interface contract"},
		{500 * time.Millisecond, "INFO", "Spinning up virtual workspace sandbox"},
		{700 * time.Millisecond, "INFO", "Generating interface definition..."},
		{500 * time.Millisecond, "INFO", "  type Handler interface { ServeHTTP(w, r) }"},
		{400 * time.Millisecond, "INFO", "SubTask [A] complete — publishing to Blackboard"},
		{350 * time.Millisecond, "INFO", "SubTask [B] received: Implement core logic"},
		{800 * time.Millisecond, "INFO", "Generating implementation with full type safety..."},
		{500 * time.Millisecond, "WARN", "Edge case detected: nil pointer in options chain"},
		{400 * time.Millisecond, "INFO", "Applying defensive guard clause — fixed"},
		{600 * time.Millisecond, "INFO", "SubTask [B] complete — 142 lines written"},
		{300 * time.Millisecond, "INFO", "Output confidence self-assessment: 0.88"},
		{200 * time.Millisecond, "INFO", "✅ Executor complete. All subtasks submitted."},
	}
	return &Executor{
		base: base{role: agent.RoleExecutor, plans: plans},
		bb:   bb,
	}
}
