package roles

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/consensus"
	"github.com/objectisnotdefined/consensus-agent/ca/pkg/llm"
)

const architectSystemPrompt = `You are a Senior Software Architect. 
Your goal is to analyze the user's task and the codebase context provided by the Navigator.
Decompose the task into a logical, step-by-step implementation plan for an AI Executor.

Your output MUST be in English and follow this structure strictly:

### Analysis
[Provide a brief technical analysis of the task and approach]

### Execution Plan
1. [Step 1: Technical instruction]
2. [Step 2: Technical instruction]
...

### Files to Modify/Create
- [File Path 1]
- [File Path 2]

Keep instructions concise, technically precise, and actionable.`

type Architect struct {
	bb       blackboard.Blackboard
	selector *llm.Selector
	status   agent.Status
	logs     []agent.LogEntry
	msgChan  chan tea.Msg
}

func NewArchitect(bb blackboard.Blackboard, selector *llm.Selector) *Architect {
	return &Architect{
		bb:       bb,
		selector: selector,
		status:   agent.StatusIdle,
		msgChan:  make(chan tea.Msg, 100),
	}
}

func (a *Architect) Role() agent.Role       { return agent.RoleArchitect }
func (a *Architect) Status() agent.Status   { return a.status }
func (a *Architect) Logs() []agent.LogEntry { return a.logs }

func (a *Architect) Reset() {
	a.status = agent.StatusIdle
	a.logs = nil
	for len(a.msgChan) > 0 {
		<-a.msgChan
	}
}

func (a *Architect) Next() tea.Cmd {
	return func() tea.Msg {
		return <-a.msgChan
	}
}

func (a *Architect) Start(ctx context.Context, task string, workspace string) tea.Cmd {
	a.status = agent.StatusRunning

	go func() {
		a.log("INFO", "🏗 Starting architectural analysis...")

		// 1. Collect Context from Blackboard
		recon, _ := a.bb.Get("codebase_summary") // Key published by Navigator
		contextStr := fmt.Sprintf("Workspace: %s\nCodebase Summary: %v\nUser Task: %s", workspace, recon, task)

		// 2. Select Model
		client, err := a.selector.SelectByRole("Architect")
		if err != nil {
			a.log("WARN", "No specialized model for Architect, using fallback.")
		} else {
			a.log("INFO", fmt.Sprintf("Using model: %s (%s)", client.Name(), client.Provider()))
		}

		// 3. Call LLM (In Phase 1, we simulate the streaming logs but capture the final result)
		// TODO: Implement real streaming LLM call in pkg/llm to pipe into a.log
		a.log("INFO", "Decomposing task into implementation steps...")
		
		// Simulated processing delay
		time.Sleep(1 * time.Second)

		// Call LLM
		resp, err := client.Chat(ctx, llm.CompletionRequest{
			Messages: []llm.Message{
				{Role: llm.RoleSystem, Content: architectSystemPrompt},
				{Role: llm.RoleUser, Content: contextStr},
			},
			Temperature: 0.2,
		})

		if err != nil {
			a.log("ERROR", fmt.Sprintf("Architect failed: %v", err))
			a.emitStatus(agent.StatusFailed)
			return
		}

		// 4. Record reasoning and plan to Blackboard
		a.log("INFO", "Plan generated. Parsing file list...")
		
		// Simple parser for Markdown structure
		plan := resp.Content
		files := extractFiles(plan)

		a.bb.Set(consensus.KeyArchitectPlan, plan)
		a.bb.Set(consensus.KeyArchitectFiles, strings.Join(files, ","))
		a.bb.Set(consensus.KeyArchitectPlanSummary, "Strategic decomposition complete.")

		a.log("INFO", fmt.Sprintf("Identified %d files to touch.", len(files)))
		a.log("INFO", "✅ Architectural plan published to Blackboard.")
		a.emitStatus(agent.StatusDone)
	}()

	return a.Next()
}

func (a *Architect) log(level, msg string) {
	entry := agent.LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
	}
	a.logs = append(a.logs, entry)
	a.msgChan <- agent.LogMsg{Role: a.Role(), Entry: entry}
}

func (a *Architect) emitStatus(s agent.Status) {
	a.status = s
	a.msgChan <- agent.StatusMsg{Role: a.Role(), Status: s}
}

// extractFiles looks for the "### Files to Modify/Create" section and extracts paths
func extractFiles(content string) []string {
	lines := strings.Split(content, "\n")
	var files []string
	inSection := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### Files") {
			inSection = true
			continue
		}
		if inSection && strings.HasPrefix(trimmed, "- ") {
			path := strings.TrimPrefix(trimmed, "- ")
			files = append(files, strings.TrimSpace(path))
		} else if inSection && trimmed == "" {
			continue
		} else if inSection && strings.HasPrefix(trimmed, "###") {
			break
		}
	}
	return files
}
