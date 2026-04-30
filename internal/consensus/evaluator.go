package consensus

import (
	"fmt"
	"strings"

	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
)

// Signal is a single quality dimension in the consensus score.
type Signal struct {
	Name   string
	Score  float64 // 0.0 ~ 1.0
	Weight float64
}

// Weighted returns the weighted contribution of this signal.
func (s Signal) Weighted() float64 { return s.Score * s.Weight }

// Report is the full result of one consensus evaluation round.
type Report struct {
	Signals    []Signal
	FinalScore float64
	Passed     bool
	Round      int    // which debate round produced this report (1-indexed)
	Critique   string // populated when Passed == false
}

// Evaluator reads quality signals from the Blackboard and computes a
// consensus score using the MCDD scoring formula:
//
//	Score = (SemanticAgreement × 0.4) + (TestPassRate × 0.4) + (SASTPassRate × 0.2)
type Evaluator struct {
	bb        blackboard.Blackboard
	threshold float64
	maxRounds int
	round     int // number of Evaluate() calls so far
}

// NewEvaluator creates an Evaluator backed by the given Blackboard.
func NewEvaluator(bb blackboard.Blackboard, threshold float64, maxRounds int) *Evaluator {
	return &Evaluator{
		bb:        bb,
		threshold: threshold,
		maxRounds: maxRounds,
	}
}

// Evaluate reads the three quality signals from the Blackboard, computes the
// weighted score, and returns a Report. The internal round counter increments
// on each call.
func (e *Evaluator) Evaluate() *Report {
	e.round++

	semantic := readFloat(e.bb, KeySemanticAgreement, 0)
	testPass := readFloat(e.bb, KeyTestPassRate, 0)
	sast := readFloat(e.bb, KeySASTPassRate, 0)

	signals := []Signal{
		{Name: "Semantic Agreement", Score: semantic, Weight: 0.4},
		{Name: "Test Pass Rate", Score: testPass, Weight: 0.4},
		{Name: "SAST Pass Rate", Score: sast, Weight: 0.2},
	}

	score := 0.0
	for _, s := range signals {
		score += s.Weighted()
	}

	passed := score >= e.threshold
	critique := ""
	if !passed {
		critique = e.buildCritique(signals)
		// Publish critique to Blackboard so agents can read it in the next debate round
		e.bb.Set(KeyDebateCritique, critique)
	}

	return &Report{
		Signals:    signals,
		FinalScore: score,
		Passed:     passed,
		Round:      e.round,
		Critique:   critique,
	}
}

// ShouldEscalate returns true when the max number of debate rounds has been
// reached without the score crossing the threshold.
func (e *Evaluator) ShouldEscalate(report *Report) bool {
	return report.Round >= e.maxRounds
}

// Reset clears the round counter, ready for a new session.
func (e *Evaluator) Reset() {
	e.round = 0
}

// Threshold returns the configured passing threshold.
func (e *Evaluator) Threshold() float64 { return e.threshold }

// ── internal helpers ─────────────────────────────────────────────────────────

// buildCritique generates a human-readable critique listing all failing signals.
func (e *Evaluator) buildCritique(signals []Signal) string {
	var parts []string
	for _, s := range signals {
		if s.Score < e.threshold {
			parts = append(parts, fmt.Sprintf(
				"%s score too low (%.2f < %.2f)", s.Name, s.Score, e.threshold,
			))
		}
	}
	if len(parts) == 0 {
		return fmt.Sprintf("Overall score below threshold %.2f.", e.threshold)
	}
	return strings.Join(parts, "; ")
}

// readFloat safely reads a float64 from the Blackboard, returning fallback on miss or type mismatch.
func readFloat(bb blackboard.Blackboard, key string, fallback float64) float64 {
	val, ok := bb.Get(key)
	if !ok {
		return fallback
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}
	return fallback
}
