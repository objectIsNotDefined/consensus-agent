package roles

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
	"github.com/objectisnotdefined/consensus-agent/ca/pkg/llm"
)

type Navigator struct {
	bb       blackboard.Blackboard
	selector *llm.Selector
	status   agent.Status
	logs     []agent.LogEntry
	msgChan  chan tea.Msg
}

func NewNavigator(bb blackboard.Blackboard, selector *llm.Selector) *Navigator {
	return &Navigator{
		bb:       bb,
		selector: selector,
		status:   agent.StatusIdle,
		msgChan:  make(chan tea.Msg, 100), // Buffer messages
	}
}

func (n *Navigator) Role() agent.Role      { return agent.RoleNavigator }
func (n *Navigator) Status() agent.Status  { return n.status }
func (n *Navigator) Logs() []agent.LogEntry { return n.logs }

func (n *Navigator) Reset() {
	n.status = agent.StatusIdle
	n.logs = nil
	// Clear channel
	for len(n.msgChan) > 0 {
		<-n.msgChan
	}
}

func (n *Navigator) Next() tea.Cmd {
	return func() tea.Msg {
		return <-n.msgChan
	}
}

func (n *Navigator) Start(ctx context.Context, task string, workspace string) tea.Cmd {
	n.status = agent.StatusRunning
	
	// Start work in background
	go func() {
		n.log("INFO", "🔍 Starting workspace reconnaissance...")
		
		// 1. Scan filesystem
		files, err := n.scanWorkspace(workspace)
		if err != nil {
			n.log("ERROR", fmt.Sprintf("Scan failed: %v", err))
			n.emitStatus(agent.StatusFailed)
			return
		}
		n.log("INFO", fmt.Sprintf("Files discovered: %d", len(files)))

		// 2. Select model
		client, err := n.selector.SelectByRole("Navigator")
		if err != nil {
			n.log("WARN", "No specialized model for Navigator, using fallback.")
			// (Assuming selector logic provides a fallback)
		} else {
			n.log("INFO", fmt.Sprintf("Using model: %s (%s)", client.Name(), client.Provider()))
		}

		// (Simulation of real LLM logic for now)
		time.Sleep(500 * time.Millisecond)
		n.log("INFO", "Building project context map...")
		
		summary := fmt.Sprintf("Project at %s has %d files.", workspace, len(files))
		n.bb.Set("codebase_summary", summary)
		
		n.log("INFO", "✅ Intelligence report published to Blackboard.")
		n.emitStatus(agent.StatusDone)
	}()

	// Return first message trigger
	return n.Next()
}

func (n *Navigator) log(level, msg string) {
	entry := agent.LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
	}
	n.logs = append(n.logs, entry)
	n.msgChan <- agent.LogMsg{Role: n.Role(), Entry: entry}
}

func (n *Navigator) emitStatus(s agent.Status) {
	n.status = s
	n.msgChan <- agent.StatusMsg{Role: n.Role(), Status: s}
}

func (n *Navigator) scanWorkspace(root string) ([]string, error) {
	var files []string
	ignoreDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		"bin":          true,
		"dist":         true,
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if ignoreDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		
		rel, _ := filepath.Rel(root, path)
		files = append(files, rel)
		return nil
	})

	return files, err
}
