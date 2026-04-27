# Project Structure Design

> This document defines the complete directory architecture for `consensus-agent`.
> Design principle: **Phase 0 is immediately buildable; Phase 1/2 packages are pre-reserved and require no directory restructuring — only implementation.**

---

## Architecture Layers

```
┌──────────────────────────────────────────────────────┐
│                   Presentation Layer                  │
│              internal/tui  ·  internal/cli            │
├──────────────────────────────────────────────────────┤
│                   Orchestration Layer                 │
│         internal/agent  ·  internal/dag              │
├──────────────────────────────────────────────────────┤
│                   Consensus Layer                     │
│                  internal/consensus                   │
├──────────────────────────────────────────────────────┤
│                 Shared State Layer (Blackboard)       │
│                 internal/blackboard                   │
├──────────────────────────────────────────────────────┤
│                   Tool Execution Layer (MCP)          │
│                    internal/mcp                       │
├──────────────────────────────────────────────────────┤
│               Infrastructure Layer                    │
│   pkg/llm  ·  pkg/config  ·  internal/memory         │
│   internal/breaker  ·  internal/cost  ·  internal/obs│
└──────────────────────────────────────────────────────┘
```

> **Key design**: Blackboard sits **above** MCP — a core principle from the README.
> Agent-to-agent communication flows through the Blackboard; tool execution (file I/O, shell, search) flows through MCP. The two layers are fully decoupled, preventing race conditions on the filesystem.

---

## Full Directory Structure

```
consensus-agent/
│
├── cmd/
│   └── ca/
│       └── main.go                        # Entry point — calls cli.Execute() only
│
├── internal/
│   │
│   ├── cli/                               # [LAYER: Presentation]
│   │   ├── root.go                        # cobra root command: ca [workspace-path]
│   │   └── completion.go                  # Shell completion (bash/zsh/fish)
│   │
│   ├── tui/                               # [LAYER: Presentation] Bubble Tea TUI
│   │   ├── app.go                         # Root Model — owns and coordinates all sub-components
│   │   ├── header.go                      # Top bar: version, workspace path, elapsed timer
│   │   ├── footer.go                      # Bottom bar: context-sensitive keybinding hints
│   │   ├── overview.go                    # Left panel (30%): agent statuses + score bars
│   │   ├── detail.go                      # Right panel (70%): scrollable viewport log stream
│   │   ├── banner.go                      # Full-screen consensus reached / failed banner
│   │   └── styles/
│   │       └── styles.go                  # All lipgloss style definitions (colors, borders, typography)
│   │
│   ├── agent/                             # [LAYER: Orchestration] Agent definitions and lifecycle
│   │   ├── types.go                       # Core types: Role, Status, LogEntry, Agent interface
│   │   ├── registry.go                    # AgentRegistry: instantiates and manages all agents
│   │   │
│   │   ├── mock/                          # [Phase 0] Mock implementations — no LLM calls
│   │   │   ├── base.go                    # Shared base: drives the log-emission chain via tea.Tick
│   │   │   ├── navigator.go               # Simulates workspace indexing
│   │   │   ├── architect.go               # Simulates task decomposition and consensus arbitration
│   │   │   ├── executor.go                # Simulates code generation
│   │   │   └── validator.go               # Simulates code review + SAST toolchain
│   │   │
│   │   └── roles/                         # [Phase 1] Real LLM-backed agent implementations
│   │       ├── navigator.go               # Large-context model: semantic indexing of the full repo
│   │       ├── architect.go               # Reasoning model: task decomposition + conflict arbitration
│   │       ├── executor.go                # Code-generation model: business logic implementation
│   │       └── validator.go               # Review model + SAST toolchain integration
│   │
│   ├── consensus/                         # [LAYER: Consensus] [Phase 1]
│   │   ├── engine.go                      # Main consensus loop: Architect cross-compares Executor + Validator output
│   │   ├── score.go                       # Confidence Score algorithm (semantic similarity, AST diff, test pass rate)
│   │   └── debate.go                      # Debate retry loop; escalates to Human-in-the-Loop after max_rounds
│   │
│   ├── dag/                               # [LAYER: Orchestration] [Phase 1]
│   │   ├── graph.go                       # DAG task graph definition (nodes, edges, dependencies)
│   │   ├── node.go                        # Node types: Task, Status, input/output schemas
│   │   └── executor.go                    # DAG executor: topological sort + parallel subtask scheduling
│   │
│   ├── blackboard/                        # [LAYER: Shared State]
│   │   ├── blackboard.go                  # Interface + in-memory implementation (sync.Map) [Phase 0]
│   │   ├── sqlite.go                      # SQLite persistence backend [Phase 1]
│   │   └── pubsub.go                      # Pub/sub subscription layer for agents to react to state changes [Phase 1]
│   │
│   ├── mcp/                               # [LAYER: Tool Execution] Model Context Protocol
│   │   ├── client.go                      # MCP client interface [Phase 1]
│   │   ├── tools.go                       # Tool registry: list of available MCP tools
│   │   └── adapters/                      # Concrete tool adapters [Phase 1]
│   │       ├── filesystem.go              # File read/write tools
│   │       ├── shell.go                   # Shell command execution tools
│   │       └── search.go                  # Code and documentation search tools
│   │
│   ├── workspace/                         # [Phase 1] Virtual sandbox environment
│   │   ├── sandbox.go                     # Isolated code execution (prevents polluting the local filesystem)
│   │   └── diff.go                        # File diff preview shown to user before any write-to-disk
│   │
│   ├── breaker/                           # [Phase 1] Circuit Breaker
│   │   └── breaker.go                     # Per-model circuit breaker (error rate + P95 latency thresholds)
│   │
│   ├── memory/                            # [Phase 2] RAG + Skill Library
│   │   ├── store.go                       # Vector DB interface (Chroma / Weaviate adapters)
│   │   ├── recall.go                      # RAG retrieval: injects past solutions into Architect context
│   │   └── decay.go                       # Memory decay policy: prunes stale or low-quality entries
│   │
│   ├── cost/                              # [Phase 2] Cost Governor
│   │   ├── governor.go                    # Global token budget manager
│   │   ├── queue.go                       # Model invocation priority queue (cheaper models for low-risk tasks)
│   │   └── dryrun.go                      # Dry-run mode: estimates cost before execution
│   │
│   └── obs/                               # [Phase 2] Observability
│       ├── logger.go                      # Structured logging per agent turn (role, model, tokens, score, latency)
│       ├── metrics.go                     # Prometheus / OpenTelemetry metrics export
│       └── replay.go                      # Session replay mode for debugging consensus parameters
│
├── pkg/                                   # Exported packages — safe to import from outside the module
│   ├── config/
│   │   └── config.go                      # Config struct + Viper loading (YAML / ENV / flag precedence)
│   │
│   └── llm/                               # [Phase 1] LLM client abstraction layer
│       ├── client.go                      # LLMClient interface + CapabilityProfile definition
│       ├── selector.go                    # Dynamic model selector: matches role profile to best available model
│       ├── openai.go                      # OpenAI / OpenAI-compatible API adapter
│       ├── anthropic.go                   # Anthropic Claude API adapter
│       └── gemini.go                      # Google Gemini API adapter
│
├── configs/
│   └── ca.yaml.example                    # Annotated configuration example
│
├── .design/
│   ├── phase0.md                          # Phase 0 detailed design
│   └── structure.md                       # ← This file: project architecture design
│
├── go.mod                                 # module github.com/objectisnotdefined/consensus-agent/ca
├── go.sum
├── Makefile
├── TODO.md
└── README.md
```

---

## Package Responsibilities

### Phase 0 — Implemented

| Package | Responsibility | Key Files |
|---|---|---|
| `internal/cli` | Cobra command parsing; launches TUI | `root.go` |
| `internal/tui` | Full-screen Bubble Tea TUI; layout management | `app.go`, `overview.go`, `detail.go` |
| `internal/agent` | Agent type definitions + Mock implementations | `types.go`, `mock/*.go` |
| `internal/blackboard` | In-memory shared state (sync.Map) | `blackboard.go` |
| `pkg/config` | YAML config loading via Viper | `config.go` |

### Phase 1 — Directory Reserved, Awaiting Implementation

| Package | Responsibility |
|---|---|
| `internal/agent/roles` | Real LLM agents (replaces mock layer) |
| `internal/consensus` | Consensus evaluation engine + Confidence Score |
| `internal/dag` | DAG task graph + parallel subtask scheduler |
| `internal/blackboard/sqlite.go` | SQLite persistence for the Blackboard |
| `internal/mcp` | MCP tool execution layer |
| `internal/workspace` | Virtual sandbox + diff preview |
| `internal/breaker` | Per-model Circuit Breaker |
| `pkg/llm` | LLM client abstractions + dynamic model selector |

### Phase 2 — Directory Reserved, Awaiting Implementation

| Package | Responsibility |
|---|---|
| `internal/memory` | Vector DB + RAG-powered skill library |
| `internal/cost` | Cost Governor + per-session token budget |
| `internal/obs` | Prometheus metrics + structured logging + replay mode |

---

## Key Design Decisions

### 1. `mock/` and `roles/` as Parallel Implementations

```
internal/agent/
  ├── mock/    ← Phase 0: pure simulation, no external dependencies, compiles instantly
  └── roles/   ← Phase 1: real LLM calls, introduces SDK dependencies
```

Both implement the same `Agent` interface. Transitioning from Phase 0 → Phase 1 requires only changing the instantiation source in `registry.go`. **The TUI layer is completely unaware of the change.**

### 2. Blackboard Sits Above MCP

The core architectural principle from the README:

- **Blackboard** = inter-agent communication bus (messages, state, shared context)
- **MCP** = tool execution channel (file reads/writes, shell commands, search)

Keeping them separate eliminates race conditions in concurrent multi-agent workflows.

### 3. CapabilityProfile-Driven Model Selection (`pkg/llm`)

```go
// Phase 1: each role declares the capabilities it requires
type CapabilityProfile struct {
    MinContextWindow int    // e.g. 1_000_000 tokens for Navigator
    ReasoningTier    string // "high" | "medium" | "low"
    SpeedTier        string // "fast" | "balanced" | "quality"
}
```

`selector.go` matches profiles to available models at runtime. **No role is bound to a specific model name**, making the system future-proof and API-key agnostic.

### 4. Centralised Styles (`internal/tui/styles/styles.go`)

All lipgloss colour constants and style definitions are declared in one place. Components consume `styles.X` — no ad-hoc colour literals scattered across files. This enables future theme switching with a single-file change.

---

## Dependency Graph

```
cmd/ca/main.go
    └── internal/cli/root.go
            └── internal/tui/app.go
                    ├── internal/tui/overview.go
                    ├── internal/tui/detail.go
                    ├── internal/tui/header.go
                    ├── internal/tui/footer.go
                    └── internal/agent/registry.go
                            ├── internal/agent/mock/navigator.go   (Phase 0)
                            ├── internal/agent/mock/architect.go   (Phase 0)
                            ├── internal/agent/mock/executor.go    (Phase 0)
                            ├── internal/agent/mock/validator.go   (Phase 0)
                            └── internal/blackboard/blackboard.go

pkg/config/config.go  ← loaded by cli/root.go; injected into tui.New()
```

> From Phase 1 onward, `registry.go` will also import `pkg/llm`, `internal/consensus`, and `internal/dag`. The TUI layer requires zero changes.
