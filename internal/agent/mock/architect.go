package mock

import (
	"time"

	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
)

// Architect simulates the Architect role: decomposing the task into a DAG of
// subtasks, orchestrating Executor and Validator, and driving the consensus loop.
type Architect struct {
	base
	bb *blackboard.Blackboard
}

// NewArchitect returns a mock Architect agent.
func NewArchitect(bb *blackboard.Blackboard) agent.Agent {
	plans := []logPlan{
		{300 * time.Millisecond, "INFO", "Reading task description from Blackboard"},
		{600 * time.Millisecond, "INFO", "Loading codebase context from Navigator index"},
		{800 * time.Millisecond, "INFO", "Task complexity assessment: HIGH"},
		{500 * time.Millisecond, "INFO", "Decomposing task into dependency graph..."},
		{400 * time.Millisecond, "INFO", "  → SubTask [A]: Define interface contract"},
		{300 * time.Millisecond, "INFO", "  → SubTask [B]: Implement core logic  (depends: A)"},
		{300 * time.Millisecond, "INFO", "  → SubTask [C]: Wire into service layer (depends: B)"},
		{500 * time.Millisecond, "INFO", "Dispatching SubTask [A] to Executor"},
		{400 * time.Millisecond, "INFO", "Dispatching test specification to Validator"},
		{700 * time.Millisecond, "INFO", "Awaiting parallel outputs from Executor & Validator..."},
		{900 * time.Millisecond, "WARN", "Confidence score below threshold: 0.72 < 0.85"},
		{400 * time.Millisecond, "INFO", "Initiating debate round 2 — requesting revision"},
		{800 * time.Millisecond, "INFO", "Executor revised output received"},
		{600 * time.Millisecond, "INFO", "Validator re-audit complete"},
		{400 * time.Millisecond, "INFO", "Consensus score: 0.91 ✅  (threshold: 0.85)"},
		{300 * time.Millisecond, "INFO", "✅ Architect complete. Merged output ready."},
	}
	return &Architect{
		base: base{role: agent.RoleArchitect, plans: plans},
		bb:   bb,
	}
}
