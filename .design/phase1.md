# Phase 1: Core Architecture & Consensus Engine

> **Goal:** Transform the Phase 0 skeleton into a functional AI engineering engine. Phase 1 focuses on real LLM integration, a persistent shared state (Blackboard), DAG-based task orchestration, and the core Consensus Protocol.

---

## 1. Core Architecture Refinement

### Capability-Based Model Selection (`pkg/llm`)
Instead of hardcoding models (e.g., "gpt-4o"), agents request a `CapabilityProfile`. This ensures the system is future-proof.

```go
type CapabilityProfile struct {
    MinContextWindow int    // e.g., 128k, 1M
    ReasoningTier    int    // 1 (Basic) to 5 (Frontier/o1)
    SpeedTier        int    // 1 (Batch) to 5 (Real-time)
}
```

- **Navigator:** High Context Window (1M+)
- **Architect:** High Reasoning Tier (Tier 5)
- **Executor:** Balanced Reasoning & Speed
- **Validator:** High Reasoning (for Audit) + Fast (for SAST)

### Persistent Blackboard (`internal/blackboard`)
Upgrade from `sync.Map` to **SQLite**. 
- **Why:** Enables session recovery, auditing, and **multi-turn dialogue context**.
- **Schema:**
    - `sessions`: Tracks workspace and overall conversation state.
    - `turns`: Records each user prompt and the final system response.
    - `tasks`: Links specific agent actions to a particular turn.
    - `artifacts`: Versioned code snippets and diffs.
- **Pub/Sub:** Agents subscribe to specific "keys" or "topics" on the blackboard to react to state changes.

---

## 2. Orchestration & DAG (`internal/dag`)

Tasks are no longer linear or random. The **Architect** generates a Directed Acyclic Graph (DAG) of subtasks.

### Workflow Example:
1. **Navigator (Node A):** Indexing & Context Gathering.
2. **Architect (Node B):** Plan Generation & DAG Construction.
3. **Executor (Node C) & Validator (Node D):** Parallel Coding and Test/Audit Generation.
4. **Consensus (Node E):** Evaluation and Merge.

---

## 3. The Consensus Protocol (`internal/consensus`)

This is the "Secret Sauce" of the project.

### Confidence Score Formula (v1)
`Score = (SemanticAgreement * 0.4) + (TestPassRate * 0.4) + (SASTPassRate * 0.2)`

1. **Semantic Agreement:** Architect compares Executor's code against the original plan.
2. **Test Pass Rate:** Code is run against Validator-generated tests in a Sandbox.
3. **SAST Pass Rate:** Output of `golangci-lint` and `gosec`.

### The Debate Loop
If `Score < Threshold` (e.g., 0.85):
- Architect generates a "Critique" and posts it to the Blackboard.
- Executor receives the critique and submits a "Revision".
- Validator re-evaluates.
- **Limit:** Max 3 rounds before **Human-in-the-Loop** escalation.

---

## 4. Virtual Workspace (`internal/workspace`)

To prevent corrupting the user's repo during "Debate":
- All code generation happens in a `tmp` directory (Sandbox).
- Tools (Go compiler, Linters) run inside this Sandbox.
- Final consensus-approved code is presented as a **Diff** in the TUI.

---

## Phase 1 Todo List

| ID | Task | Key Points |
|:---|:---|:---|
| **1.1** | **Visual Config Portal** | Create a simple HTML/Web-based UI to manage `ca.yaml`, API keys, and model capability overrides. |
| **1.2** | **LLM Provider Layer** | Implement `pkg/llm` with support for OpenAI and Anthropic. Add the `Selector` logic to match `CapabilityProfile` to config. |
| **1.3** | **SQLite Blackboard** | Implement `internal/blackboard/sqlite.go`. Define schema for `sessions`, `turns`, `tasks`, and `artifacts` to support multi-turn memory. |
| **1.4** | **DAG Executor** | Create a basic engine in `internal/dag` that can run tasks in parallel based on dependencies. |
| **1.5** | **Navigator Real Scan** | Replace Navigator mock with real logic that traverses the filesystem and builds a "Map" for the LLM context. |
| **1.6** | **Consensus Evaluator** | Implement the first version of the scoring algorithm in `internal/consensus`. |
| **1.7** | **SAST Integration** | Hook up `golangci-lint` and `gosec` inside the `Validator` role. |
| **1.8** | **Sandbox & Diff** | Implement the `internal/workspace` logic to copy files to a temporary area and generate `git diff` output. |
| **1.9** | **Human-in-the-Loop** | Add a TUI state that pauses execution and prompts the user for "Approve / Reject / Comment" when consensus fails. |
| **1.10** | **Circuit Breaker** | Implement `internal/breaker` to track API failures and switch models if one provider is down. |

---

## Success Criteria for Phase 1
- [ ] A user can type a prompt (e.g., "Add a GET /health endpoint").
- [ ] Real LLM calls are made to at least two different models.
- [ ] Code is validated by `golangci-lint`.
- [ ] The TUI shows a diff of the changes.
- [ ] The user can approve the change to write it to disk.
- [ ] Config can be managed via a visual/interactive interface instead of raw YAML editing.
