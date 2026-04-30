package roles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/consensus"
	"github.com/objectisnotdefined/consensus-agent/ca/pkg/llm"
)

const executorSystemPrompt = `You are an Expert Software Engineer. 
Your task is to implement the changes described in the provided Architectural Plan.
You will be given the plan and the current content of relevant files.

CRITICAL INSTRUCTIONS:
1. Output ONLY the files that need to be modified or created.
2. For each file, use the following format:
FILE: [relative path]
[full code content]

3. Do not include conversational text or explanations outside the FILE format.
4. Ensure the code is production-ready, follows best practices, and matches the plan perfectly.`

type Executor struct {
	bb       blackboard.Blackboard
	selector *llm.Selector
	status   agent.Status
	logs     []agent.LogEntry
	msgChan  chan tea.Msg
}

func NewExecutor(bb blackboard.Blackboard, selector *llm.Selector) *Executor {
	return &Executor{
		bb:       bb,
		selector: selector,
		status:   agent.StatusIdle,
		msgChan:  make(chan tea.Msg, 100),
	}
}

func (e *Executor) Role() agent.Role       { return agent.RoleExecutor }
func (e *Executor) Status() agent.Status   { return e.status }
func (e *Executor) Logs() []agent.LogEntry { return e.logs }

func (e *Executor) Reset() {
	e.status = agent.StatusIdle
	e.logs = nil
	for len(e.msgChan) > 0 {
		<-e.msgChan
	}
}

func (e *Executor) Next() tea.Cmd {
	return func() tea.Msg {
		return <-e.msgChan
	}
}

func (e *Executor) Start(ctx context.Context, task string, workspace string) tea.Cmd {
	e.status = agent.StatusRunning

	go func() {
		e.log("INFO", "⚙️ Starting code implementation...")

		// 1. Collect Input from Blackboard
		plan, _ := e.bb.Get(consensus.KeyArchitectPlan)
		sbPathVal, _ := e.bb.Get(consensus.KeySandboxPath)
		sbPath := sbPathVal.(string)

		e.log("INFO", "Reading files and architectural plan...")

		// 2. Prepare Context (Plan + File Contents)
		// For Phase 1, we pass the plan and let the LLM decide.
		// In a more advanced version, we'd read the specific files listed by the Architect.
		contextStr := fmt.Sprintf("Architectural Plan:\n%v\n\nTask: %s", plan, task)

		// 3. Select Model
		client, err := e.selector.SelectByRole("Executor")
		if err != nil {
			e.log("WARN", "No specialized model for Executor, using fallback.")
		} else {
			e.log("INFO", fmt.Sprintf("Using model: %s (%s)", client.Name(), client.Provider()))
		}

		// 4. Call LLM
		e.log("INFO", "Generating code in sandbox...")
		resp, err := client.Chat(ctx, llm.CompletionRequest{
			Messages: []llm.Message{
				{Role: llm.RoleSystem, Content: executorSystemPrompt},
				{Role: llm.RoleUser, Content: contextStr},
			},
			Temperature: 0.1, // Low temperature for code generation consistency
		})

		if err != nil {
			e.log("ERROR", fmt.Sprintf("Code generation failed: %v", err))
			e.emitStatus(agent.StatusFailed)
			return
		}

		// 5. Parse and Write Files to Sandbox
		filesWritten := e.parseAndWrite(resp.Content, sbPath)
		
		if filesWritten == 0 {
			e.log("WARN", "No files were extracted from LLM response. Check format.")
		} else {
			e.log("INFO", fmt.Sprintf("Successfully updated %d files in sandbox.", filesWritten))
		}

		e.bb.Set(consensus.KeyExecutorCode, resp.Content)
		e.bb.Set(consensus.KeySemanticAgreement, 0.9) // Self-assessment

		e.log("INFO", "✅ Implementation phase complete.")
		e.emitStatus(agent.StatusDone)
	}()

	return e.Next()
}

func (e *Executor) parseAndWrite(content string, sandboxPath string) int {
	parts := strings.Split(content, "FILE: ")
	count := 0

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		lines := strings.Split(part, "\n")
		if len(lines) < 2 {
			continue
		}

		relPath := strings.TrimSpace(lines[0])
		code := strings.Join(lines[1:], "\n")
		
		// Clean up code blocks if LLM included them
		code = strings.TrimPrefix(code, "```")
		code = strings.TrimPrefix(code, "go") // handle ```go
		code = strings.TrimSuffix(code, "```")
		code = strings.TrimSpace(code)

		absPath := filepath.Join(sandboxPath, relPath)
		
		// Ensure directory exists
		os.MkdirAll(filepath.Dir(absPath), 0755)

		err := os.WriteFile(absPath, []byte(code), 0644)
		if err != nil {
			e.log("ERROR", fmt.Sprintf("Failed to write %s: %v", relPath, err))
			continue
		}
		
		e.log("INFO", fmt.Sprintf("  → Written: %s", relPath))
		count++
	}
	return count
}

func (e *Executor) log(level, msg string) {
	entry := agent.LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
	}
	e.logs = append(e.logs, entry)
	e.msgChan <- agent.LogMsg{Role: e.Role(), Entry: entry}
}

func (e *Executor) emitStatus(s agent.Status) {
	e.status = s
	e.msgChan <- agent.StatusMsg{Role: e.Role(), Status: s}
}
