package mock

import (
	"time"

	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
)

// Validator simulates the Validator role: running dual-track validation —
// LLM-based semantic code review combined with SAST toolchain (golangci-lint, gosec).
type Validator struct {
	base
	bb *blackboard.Blackboard
}

// NewValidator returns a mock Validator agent.
func NewValidator(bb *blackboard.Blackboard) agent.Agent {
	plans := []logPlan{
		{400 * time.Millisecond, "INFO", "Awaiting code output from Executor..."},
		{600 * time.Millisecond, "INFO", "Code received — starting dual-track validation"},
		{400 * time.Millisecond, "INFO", "Track 1 [SAST]: Running golangci-lint..."},
		{800 * time.Millisecond, "INFO", "  golangci-lint: 0 issues found  [PASS ✅]"},
		{350 * time.Millisecond, "INFO", "Track 1 [SAST]: Running gosec security scan..."},
		{700 * time.Millisecond, "INFO", "  gosec: 0 HIGH  0 MEDIUM  0 LOW  [PASS ✅]"},
		{500 * time.Millisecond, "INFO", "Track 2 [LLM]:  Semantic review in progress..."},
		{900 * time.Millisecond, "WARN", "Semantic issue: missing error propagation in handler"},
		{400 * time.Millisecond, "INFO", "Flagging for Architect review — revision requested"},
		{600 * time.Millisecond, "INFO", "Revised code received — re-running validation"},
		{600 * time.Millisecond, "INFO", "  golangci-lint: 0 issues found  [PASS ✅]"},
		{400 * time.Millisecond, "INFO", "  Semantic review score: 0.94"},
		{300 * time.Millisecond, "INFO", "Combined validation score: 0.91"},
		{200 * time.Millisecond, "INFO", "✅ Validator complete. Report submitted to Architect."},
	}
	return &Validator{
		base: base{role: agent.RoleValidator, plans: plans},
		bb:   bb,
	}
}
