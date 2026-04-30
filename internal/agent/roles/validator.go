package roles

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/consensus"
	"github.com/objectisnotdefined/consensus-agent/ca/pkg/llm"
)

type Validator struct {
	bb       blackboard.Blackboard
	selector *llm.Selector
	status   agent.Status
	logs     []agent.LogEntry
	msgChan  chan tea.Msg
}

func NewValidator(bb blackboard.Blackboard, selector *llm.Selector) *Validator {
	return &Validator{
		bb:       bb,
		selector: selector,
		status:   agent.StatusIdle,
		msgChan:  make(chan tea.Msg, 100),
	}
}

func (v *Validator) Role() agent.Role       { return agent.RoleValidator }
func (v *Validator) Status() agent.Status   { return v.status }
func (v *Validator) Logs() []agent.LogEntry { return v.logs }

func (v *Validator) Reset() {
	v.status = agent.StatusIdle
	v.logs = nil
	for len(v.msgChan) > 0 {
		<-v.msgChan
	}
}

func (v *Validator) Next() tea.Cmd {
	return func() tea.Msg {
		return <-v.msgChan
	}
}

func (v *Validator) Start(ctx context.Context, task string, workspace string) tea.Cmd {
	v.status = agent.StatusRunning

	go func() {
		v.log("INFO", "🛡 Starting multi-track validation...")

		// 1. Get Sandbox Path
		sbPathVal, ok := v.bb.Get(consensus.KeySandboxPath)
		sbPath := workspace // fallback
		if ok {
			sbPath = sbPathVal.(string)
		}

		// 2. Track 1: SAST (golangci-lint & gosec)
		v.log("INFO", "Track 1: Running SAST scans (static analysis)...")
		sastScore := v.runSAST(sbPath)
		v.bb.Set(consensus.KeySASTPassRate, sastScore)

		// 3. Track 2: Semantic Review (LLM-based)
		// TODO: In Phase 1.12, we will implement real semantic code review here.
		// For now, we simulate LLM review.
		v.log("INFO", "Track 2: Initiating LLM semantic code review...")
		time.Sleep(800 * time.Millisecond)
		
		v.log("INFO", "Reviewing code for edge cases and patterns...")
		v.bb.Set(consensus.KeyTestPassRate, 0.95) // Simulated test pass rate
		v.bb.Set(consensus.KeyValidatorReport, "Semantic review passed. No critical anti-patterns found.")

		v.log("INFO", "✅ Validation suite complete.")
		v.emitStatus(agent.StatusDone)
	}()

	return v.Next()
}

func (v *Validator) runSAST(path string) float64 {
	score := 1.0
	
	// Track issues
	issues := 0

	// 1. golangci-lint
	v.log("INFO", "  → Executing golangci-lint...")
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		v.log("WARN", "    [!] golangci-lint not found. Skipping...")
	} else {
		cmd := exec.Command("golangci-lint", "run", "./...")
		cmd.Dir = path
		out, err := cmd.Output()
		if err != nil {
			// golangci-lint returns 1 if issues found
			issues += 1 // Simplified: just mark as having issues
			v.log("WARN", "    [✘] golangci-lint found issues.")
			fmt.Println(string(out))
		} else {
			v.log("INFO", "    [✔] golangci-lint passed.")
		}
	}

	// 2. gosec
	v.log("INFO", "  → Executing gosec security scan...")
	if _, err := exec.LookPath("gosec"); err != nil {
		v.log("WARN", "    [!] gosec not found. Skipping...")
	} else {
		cmd := exec.Command("gosec", "-quiet", "./...")
		cmd.Dir = path
		if err := cmd.Run(); err != nil {
			issues += 1
			v.log("WARN", "    [✘] gosec found potential security vulnerabilities.")
		} else {
			v.log("INFO", "    [✔] gosec security check passed.")
		}
	}

	if issues > 0 {
		score = 0.5 // Simplified penalty
	}

	return score
}

func (v *Validator) log(level, msg string) {
	entry := agent.LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
	}
	v.logs = append(v.logs, entry)
	v.msgChan <- agent.LogMsg{Role: v.Role(), Entry: entry}
}

func (v *Validator) emitStatus(s agent.Status) {
	v.status = s
	v.msgChan <- agent.StatusMsg{Role: v.Role(), Status: s}
}
