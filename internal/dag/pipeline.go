package dag

import "github.com/objectisnotdefined/consensus-agent/ca/internal/agent"

// MCDDPipeline returns the standard Model Consensus Driven Development pipeline:
//
//	Navigator ──→ Architect ──→ Executor  ──→ (Consensus)
//	                        └──→ Validator ──→ (Consensus)
//
// Execution order:
//  1. Navigator  — filesystem reconnaissance, context building
//  2. Architect  — plan generation (depends on Navigator)
//  3. Executor   — code generation (depends on Architect)
//  4. Validator  — test & audit generation (depends on Architect, parallel with Executor)
func MCDDPipeline() (*DAG, error) {
	return Build([]Node{
		{
			ID:   "navigator",
			Role: agent.RoleNavigator,
			Deps: []NodeID{},
		},
		{
			ID:   "architect",
			Role: agent.RoleArchitect,
			Deps: []NodeID{"navigator"},
		},
		{
			ID:   "executor",
			Role: agent.RoleExecutor,
			Deps: []NodeID{"architect"},
		},
		{
			ID:   "validator",
			Role: agent.RoleValidator,
			Deps: []NodeID{"architect"}, // parallel with executor after architect finishes
		},
	})
}
