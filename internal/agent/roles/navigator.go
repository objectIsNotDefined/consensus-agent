package roles

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
	"github.com/objectisnotdefined/consensus-agent/ca/pkg/llm"
)

const navigatorSystemPrompt = `You are a Codebase Intelligence Navigator.
Your goal is to provide a high-level semantic map of the codebase for a Software Architect.
Analyze the provided file list and the contents of key entry-point files.

Your summary should include:
1. The primary purpose of the project.
2. The core technology stack (languages, frameworks).
3. Key architectural patterns (e.g., MVC, Hexagonal, DDD).
4. Critical components and their responsibilities.

Keep the summary technical, concise, and structured.`

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

func (n *Navigator) Role() agent.Role       { return agent.RoleNavigator }
func (n *Navigator) Status() agent.Status   { return n.status }
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

		// 2. Dynamically identify key files for context
		keyFiles := n.identifyKeyFiles(files)
		var contextContent strings.Builder
		contextContent.WriteString(fmt.Sprintf("File list:\n%s\n\n", strings.Join(files, "\n")))

		for _, kf := range keyFiles {
			absPath := filepath.Join(workspace, kf)
			if _, err := os.Stat(absPath); err == nil {
				data, err := os.ReadFile(absPath)
				if err == nil {
					n.log("INFO", fmt.Sprintf("Reading key file: %s", kf))
					contextContent.WriteString(fmt.Sprintf("--- Content of %s ---\n%s\n\n", kf, string(data)))
				}
			}
		}

		// 3. Select model
		client, err := n.selector.SelectByRole("Navigator")
		if err != nil {
			n.log("ERROR", fmt.Sprintf("No model found for Navigator: %v", err))
			n.emitStatus(agent.StatusFailed)
			return
		}
		n.log("INFO", fmt.Sprintf("Using model: %s (%s)", client.Name(), client.Provider()))

		// 4. Call LLM to generate intelligence report
		n.log("INFO", "Analyzing codebase architecture...")
		resp, err := client.Chat(ctx, llm.CompletionRequest{
			Messages: []llm.Message{
				{Role: llm.RoleSystem, Content: navigatorSystemPrompt},
				{Role: llm.RoleUser, Content: contextContent.String()},
			},
			Temperature: 0.2,
		})

		if err != nil {
			n.log("ERROR", fmt.Sprintf("Navigator analysis failed: %v", err))
			n.emitStatus(agent.StatusFailed)
			return
		}

		summary := resp.Content
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

// identifyKeyFiles uses heuristics to pick the most informative files for codebase context.
func (n *Navigator) identifyKeyFiles(files []string) []string {
	// 1. Critical configuration and documentation files
	anchors := map[string]bool{
		"README.md":        true,
		"README":           true,
		"go.mod":           true,
		"package.json":     true,
		"requirements.txt": true,
		"Cargo.toml":       true,
		"Makefile":         true,
		"docker-compose.yml": true,
	}

	// 2. Common entry point patterns
	entryPatterns := []string{"main.", "index.", "app.", "server."}

	var found []string
	seen := make(map[string]bool)

	// Search for anchors first
	for _, f := range files {
		base := filepath.Base(f)
		if anchors[base] {
			found = append(found, f)
			seen[f] = true
		}
	}

	// Search for entry points (only in root or cmd/ or src/)
	for _, f := range files {
		if seen[f] {
			continue
		}
		base := filepath.Base(f)
		dir := filepath.Dir(f)

		isEntry := false
		for _, p := range entryPatterns {
			if strings.HasPrefix(base, p) {
				isEntry = true
				break
			}
		}

		if isEntry {
			// Prioritize root or common subdirs
			if dir == "." || strings.HasPrefix(dir, "cmd") || strings.HasPrefix(dir, "src") {
				found = append(found, f)
				seen[f] = true
			}
		}
		
		// Stop if we have enough context
		if len(found) > 8 {
			break
		}
	}

	return found
}
