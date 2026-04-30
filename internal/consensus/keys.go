// Package consensus implements the Consensus Evaluator for the MCDD pipeline.
//
// Scoring formula (v1):
//
//	Score = (SemanticAgreement × 0.4) + (TestPassRate × 0.4) + (SASTPassRate × 0.2)
//
// All signal values are published to the Blackboard by their respective agents
// and read here for evaluation.
package consensus

// Blackboard KV keys used by the consensus pipeline.
// Agents MUST write these keys before emitting StatusDone.
const (
	// Published by Executor
	KeyExecutorCode      = "executor.code"      // string: generated code
	KeySemanticAgreement = "executor.semantic"  // float64: how well code follows the plan (0–1)

	// Published by Architect
	KeyArchitectPlan        = "architect.plan"    // string: structured execution plan
	KeyArchitectFiles       = "architect.files"   // string: JSON list of files to touch
	KeyArchitectThoughts    = "architect.thoughts" // string: technical reasoning
	KeyArchitectPlanSummary = "architect.summary" // string: short summary of the approach

	// Environment / Session Keys
	KeySandboxPath = "session.sandbox_path" // string: absolute path to the temp sandbox

	// Published by Validator
	KeyTestPassRate    = "validator.test_pass_rate" // float64: test pass ratio (0–1)
	KeySASTPassRate    = "validator.sast_pass_rate" // float64: SAST pass ratio (0–1)
	KeyValidatorReport = "validator.report"         // string: human-readable report

	// Written by Evaluator during a Debate round
	KeyDebateCritique = "debate.critique" // string: critique text for next revision
)
