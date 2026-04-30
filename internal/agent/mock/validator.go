package mock

import (
	"time"

	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/consensus"
)

// Validator simulates the Validator role: running dual-track validation —
// LLM-based semantic code review combined with SAST toolchain (golangci-lint, gosec).
type Validator struct {
	base
	bb blackboard.Blackboard
}

// NewValidator returns a mock Validator agent.
func NewValidator(bb blackboard.Blackboard) agent.Agent {
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
	v := &Validator{
		base: base{role: agent.RoleValidator, plans: plans},
		bb:   bb,
	}
	// onDone writes SAST and test pass rates to the Blackboard.
	// Round 1 scores combine with Executor's low semantic score to fall below
	// threshold (0.72×0.4 + 0.90×0.4 + 1.0×0.2 = 0.848 < 0.85) → triggers debate.
	// Round 2 scores push the final result above threshold.
	v.base.onDone = func() {
		_, isDebate := bb.Get(consensus.KeyDebateCritique)
		if isDebate {
			// Improved after debate round
			bb.Set(consensus.KeyTestPassRate, 0.92)
			bb.Set(consensus.KeyValidatorReport, "All tests pass. Semantic issue resolved after revision.")
		} else {
			// Initial validation
			bb.Set(consensus.KeyTestPassRate, 0.90)
			bb.Set(consensus.KeyValidatorReport, "SAST clean. Missing error propagation flagged.")
		}
		bb.Set(consensus.KeySASTPassRate, 1.0) // SAST always clean in mock
	}
	return v
}
