# Phase 0: CLI Bootstrap Design

> **Goal:** Build a fully runnable skeleton for `consensus-agent`. After running `ca`, a full-screen Bubble Tea TUI launches — showing all agents running in parallel — without any real LLM calls. Phase 0 is about framework + UX, not intelligence.

---

## 1. User Experience

### Entry Point

```bash
ca ./my-project
ca /absolute/path/to/repo
ca .
```

The only argument is a **file or directory path** — the workspace the agents will operate on. No flags required.

Once launched, the TUI takes over the full terminal. A text input at the top of the interface (similar to Codex / Claude Code) accepts the user's task description. After pressing `Enter`, all 4 agents begin running in **parallel**.

### Interaction Flow

```
Prompt> $ ca ./my-project
           │
           ▼
  ┌─── TUI launches (full screen) ────────────────────┐
  │   "What should the agents do?" [text input]        │
  │   User types: "Add JWT auth middleware"            │
  │   [Enter]                                         │
  └────────────────────────────────────────────────────┘
           │
           ▼
  All 4 agents start in parallel
  Left panel updates statuses in real time
  Right panel shows selected agent's log stream
           │
           ▼
  All agents complete → "Consensus Reached ✅" banner
           │
  [Q / Ctrl+C] → TUI exits, terminal restored
```

---

## 2. TUI Layout

Terminal dimensions are read from `tea.WindowSizeMsg` and all panels resize proportionally.

```
┌────────────────────────────────────────────────────────────────────────┐
│  ⚡ consensus-agent  v0.1.0    📁 ./my-project                  [ESC] │  ← Header
├─── OVERVIEW (30%) ─────────────┬─── AGENT DETAIL (70%) ──────────────┤
│                                │                                       │
│  AGENTS                        │  ▶ Architect                         │
│  ┌────────────────────────┐   │  ──────────────────────────────────   │
│  │ ◉ Navigator   🔄 Run  │   │  12:01:01 [INFO] Task received        │
│  │ ◉ Architect   🔄 Run  │   │  12:01:02 [INFO] Decomposing into     │
│  │ ◉ Executor    🔄 Run  │   │            3 subtasks                 │  ← Viewport
│  │ ◉ Validator   🔄 Run  │   │  12:01:03 [INFO] Subtask A → Exec     │   (scrollable)
│  └────────────────────────┘   │  12:01:04 [WARN] Confidence low...   │
│                                │  12:01:05 [INFO] Retry round 2       │
│  CONSENSUS SCORE               │                                       │
│  ████████░░  0.82 / 0.85      │                                       │
│                                │                                       │
│  TOKEN BUDGET                  │                                       │
│  ██████░░░░  6.2k / 10k       │                                       │
│                                │                                       │
├────────────────────────────────┴───────────────────────────────────────┤
│  [Tab] Switch agent  [↑↓] Scroll  [R] Replay  [Q] Quit               │  ← Footer
└────────────────────────────────────────────────────────────────────────┘
```

### Panel Breakdown

| Panel | Width | Content |
|---|---|---|
| **Header** | 100% | App name, version, workspace path, elapsed timer |
| **Overview (Left)** | 30% | 4 agent status rows + Consensus Score bar + Token Budget bar |
| **Detail (Right)** | 70% | Scrollable log viewport for the currently selected agent |
| **Footer** | 100% | Context-sensitive keybinding hints |

### Agent Status Indicators

| State | Icon | Color |
|---|---|---|
| Idle | `◯` | Dim gray |
| Running | `🔄` + spinner | Cyan |
| Done | `✅` | Green |
| Failed | `❌` | Red |
| Escalated | `⚠️` | Yellow |

### Log Level Colors

| Level | Color |
|---|---|
| `INFO` | White |
| `WARN` | Yellow |
| `ERROR` | Red / Bold |
| `DEBUG` | Dim gray |

---

## 3. Project Structure

```
consensus-agent/
│
├── cmd/
│   └── ca/
│       └── main.go                  # Minimal entry: calls cli.Execute()
│
├── internal/
│   ├── cli/
│   │   └── root.go                  # cobra root command: ca [path]
│   │
│   ├── agent/
│   │   ├── types.go                 # Agent interface, Role, Status, LogEntry
│   │   ├── mock.go                  # MockAgent: simulates work, emits fake logs
│   │   └── registry.go              # AgentRegistry: manages all 4 agent instances
│   │
│   ├── blackboard/
│   │   └── blackboard.go            # In-memory shared state (sync.Map based)
│   │
│   └── tui/
│       ├── app.go                   # Root tea.Model — owns all sub-components
│       ├── header.go                # Header bar component
│       ├── footer.go                # Footer bar component
│       ├── overview.go              # Left panel: agent list + score bars
│       └── detail.go                # Right panel: bubbles/viewport log stream
│
├── pkg/
│   └── config/
│       └── config.go                # Config struct (loaded from ca.yaml via Viper)
│
├── configs/
│   └── ca.yaml.example              # Annotated config example
│
├── .design/
│   └── phase0.md                    # ← This file
│
├── go.mod                           # module github.com/objectisnotdefined/consensus-agent/ca
├── Makefile
├── TODO.md
└── README.md
```

---

## 4. Core Types

### `internal/agent/types.go`

```go
package agent

import (
    "context"
    "time"

    tea "github.com/charmbracelet/bubbletea"
)

type Role string

const (
    RoleNavigator Role = "Navigator"
    RoleArchitect Role = "Architect"
    RoleExecutor  Role = "Executor"
    RoleValidator Role = "Validator"
)

type Status int

const (
    StatusIdle Status = iota
    StatusRunning
    StatusDone
    StatusFailed
    StatusEscalated
)

type LogEntry struct {
    Timestamp time.Time
    Level     string // "INFO" | "WARN" | "ERROR" | "DEBUG"
    Message   string
}

// LogMsg is a tea.Msg emitted by agents when a new log line is ready
type LogMsg struct {
    Role  Role
    Entry LogEntry
}

// StatusMsg is a tea.Msg emitted when an agent's status changes
type StatusMsg struct {
    Role   Role
    Status Status
}

// Agent is the minimal interface all agents must implement
type Agent interface {
    Role() Role
    Status() Status
    Logs() []LogEntry
    // Start begins the agent's work and returns a tea.Cmd that drives the TUI updates
    Start(ctx context.Context, task string, workspace string) tea.Cmd
}
```

### `internal/blackboard/blackboard.go`

```go
// Blackboard is a thread-safe in-memory shared state layer.
// Agents read/write via Get/Set. In Phase 0 this is only scaffolding.
// Phase 1 will add SQLite persistence and a pub/sub subscription model.
type Blackboard struct {
    store sync.Map
}

func (b *Blackboard) Set(key string, value any)
func (b *Blackboard) Get(key string) (any, bool)
```

---

## 5. TUI Component Design

### `internal/tui/app.go` — Root Model

```go
type Model struct {
    // sub-components
    header   header.Model
    footer   footer.Model
    overview overview.Model
    detail   detail.Model

    // state
    agents      []*agent.MockAgent
    activeAgent int    // index of selected agent (Tab cycles this)
    task        string // task string entered by user
    inputMode   bool   // true = user is typing task, false = viewing agents

    // dimensions
    width, height int
}
```

**Key message handlers:**
- `tea.WindowSizeMsg` → recalculate all panel widths/heights
- `tea.KeyMsg`:
  - `Tab` → cycle `activeAgent`, update `detail` viewport
  - `↑` / `↓` → delegate to `detail` viewport scroll
  - `R` → reset all MockAgents, restart simulation
  - `Q` / `Ctrl+C` → `tea.Quit`
  - `Enter` (input mode) → lock task, start all agents in parallel
- `agent.LogMsg` → append to correct agent's log, re-render detail if active
- `agent.StatusMsg` → update agent status in overview

### `internal/tui/overview.go` — Left Panel

Renders a fixed-width (30% of terminal) box containing:
1. `AGENTS` section: one row per agent with spinner/icon + status label
2. `CONSENSUS SCORE` section: ASCII progress bar (lipgloss)
3. `TOKEN BUDGET` section: ASCII progress bar (lipgloss)

Uses `lipgloss.NewStyle().Width(m.width).Height(m.height)` to fill exact space.

### `internal/tui/detail.go` — Right Panel

Wraps `bubbles/viewport`:
- Auto-scrolls to bottom when new `LogMsg` arrives
- Manual `↑↓` scroll when user navigates
- Renders log lines with timestamp, colored level badge, and message
- Title bar shows `▶ [Agent Role]` and line count

---

## 6. Mock Agent Behavior

All 4 MockAgents run **concurrently** from the moment the user submits their task. Each emits log lines on independent timers with randomized delays (e.g. `300ms–800ms` between entries), simulating real async work.

### Navigator Mock Sequence
```
[INFO] Scanning workspace: ./my-project
[INFO] Found 42 Go files, 3 config files
[INFO] Building semantic index...
[INFO] Index ready. Context window usage: 18%
[INFO] Publishing codebase snapshot to Blackboard
[INFO] Done ✅
```

### Architect Mock Sequence
```
[INFO] Received task from Blackboard
[INFO] Analyzing codebase context...
[INFO] Decomposing task into 3 subtasks
[INFO]   → SubTask A: Define middleware interface
[INFO]   → SubTask B: Implement JWT validation logic
[INFO]   → SubTask C: Wire into router
[INFO] Dispatching SubTask A to Executor
[INFO] Dispatching SubTask A test stubs to Validator
[INFO] Awaiting outputs...
[WARN] Confidence score below threshold: 0.72
[INFO] Initiating debate round 2
[INFO] Consensus reached: 0.88 ✅
```

### Executor Mock Sequence
```
[INFO] SubTask A received: Define middleware interface
[INFO] Generating code in virtual workspace...
[INFO] Function `NewJWTMiddleware` written
[INFO] SubTask B received
[INFO] Generating token validation logic...
[WARN] Edge case detected: expired token handling
[INFO] Handling added
[INFO] All subtasks complete, submitting output
```

### Validator Mock Sequence
```
[INFO] Received test stub request from Architect
[INFO] Generating test stubs for middleware...
[INFO] Running golangci-lint... [PASS]
[INFO] Running gosec... [PASS]
[WARN] Missing error case: malformed token
[INFO] Flagging for Architect review
[INFO] Semantic review score: 0.91
[INFO] Validation report submitted ✅
```

---

## 7. Dependencies

| Package | Purpose |
|---|---|
| `github.com/charmbracelet/bubbletea` | TUI runtime |
| `github.com/charmbracelet/bubbles` | `viewport`, `spinner`, `textinput` components |
| `github.com/charmbracelet/lipgloss` | Styling, layout, progress bars |
| `github.com/spf13/cobra` | CLI command parsing |
| `github.com/spf13/viper` | Config file loading (YAML) |

No external AI/LLM SDKs in Phase 0.

---

## 8. Configuration (`ca.yaml`)

```yaml
# consensus-agent configuration
consensus:
  threshold: 0.85       # minimum score to accept output
  max_rounds: 3         # max debate rounds before human escalation

cost:
  token_budget: 10000   # total token budget per session

# Phase 1: fill in real model API keys
models: []
```

Viper loads from `~/.config/ca/ca.yaml` (default) or `--config` flag override.

---

## 9. Makefile

```makefile
.PHONY: build install dev lint

build:
	go build -o bin/ca ./cmd/ca/

install:
	go install ./cmd/ca/
	@echo "✅ 'ca' installed to $(shell go env GOPATH)/bin"

dev:
	go run ./cmd/ca/ .

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
```

After `make install`, the `ca` command is available globally (assuming `$GOPATH/bin` is in `$PATH`).

---

## 10. What Phase 0 Does NOT Include

The following are explicitly **out of scope** for Phase 0 and belong to later phases:

| Feature | Phase |
|---|---|
| Real LLM API calls (OpenAI, Anthropic, Gemini) | Phase 1 |
| Confidence Score real computation (AST diff, semantic sim) | Phase 1 |
| Blackboard SQLite persistence | Phase 1 |
| Circuit Breaker per model | Phase 1 |
| golangci-lint / gosec real integration | Phase 1 |
| DAG task graph executor | Phase 1 |
| Vector DB + RAG skill library | Phase 2 |
| Cost Governor token tracking | Phase 2 |
| PR / commit auto-generation | Phase 2 |
| Web-based DAG visualizer | Phase 2 |
