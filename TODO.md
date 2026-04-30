# Consensus-Agent — Development Roadmap

> This document tracks the full development plan for `consensus-agent`.
> Phase 1 covers core architecture foundations. Phase 2 covers optimization and advanced capabilities.

---

## Phase 1 — Core Architecture Foundations

> Goal: Build a working, robust, and well-defined multi-model consensus pipeline.

### Architecture & Design
- [x] Define `CapabilityProfile` struct for each role (context window size, reasoning tier, speed tier)
- [x] Implement dynamic model selector: load available models from config and match to role capability profiles at runtime
- [x] Design and implement the **Blackboard** shared state layer (in-memory KV with optional SQLite persistence)
- [x] Decouple agent communication from MCP tool execution via the Blackboard abstraction

### Consensus Engine
- [x] Define the **Confidence Score** algorithm (e.g., semantic similarity, AST diff agreement, test pass rate)
- [x] Implement consensus evaluation loop: Architect cross-compares Executor output and Validator audit
- [x] Set configurable `consensus_threshold` (default: 0.85)
- [x] Implement debate retry loop: re-run disagreeing agents up to `max_rounds` (default: 3)
- [x] Implement **Human-in-the-Loop** escalation: pause and prompt developer when consensus fails after max rounds

### Resilience
- [ ] Implement **Circuit Breaker** per model integration
  - [ ] Track error rate and P95 latency per model
  - [ ] Auto-disable model and fallback to next-best when thresholds are breached
  - [ ] Implement half-open state for recovery probing
- [x] Handle API-level failures gracefully: timeouts, rate limits, malformed outputs

### Validator Revamp
- [x] Replace Copilot/Codex placeholder with LLM-powered semantic code review agent
- [x] Integrate `golangci-lint` as a mandatory validation step
- [x] Integrate `gosec` for security-focused static analysis
- [x] Combine LLM audit score + SAST pass/fail into a unified Validator report fed to the consensus engine

### CLI / TUI
- [x] Scaffold the Bubble Tea TUI shell
- [x] Implement interactive task input and live agent status display
- [ ] Add code diff preview before any file is written to disk
- [ ] Implement automated PR/commit generation after consensus is reached

### Infrastructure
- [x] Design DAG task graph schema and executor
- [x] Implement virtual workspace (isolated sandbox for code simulation and test execution)
- [x] Write configuration schema (YAML): model API keys, capability profiles, consensus thresholds, cost limits

---

## Phase 2 — Optimization & Advanced Capabilities

> Goal: Make the system smarter, more cost-efficient, and observable over time.

### Memory & Self-Evolution
- [ ] Integrate a Vector Database (e.g., Chroma or Weaviate) as the persistent skill library
- [ ] Implement storage of successful fix/refactor paths as embedded documents
- [ ] Implement RAG recall: during Architect decision-making, retrieve and inject relevant past solutions as context
- [ ] Add a "memory decay" policy to prune stale or low-quality entries

### Task Graph Visualization
- [ ] Build a web-based or TUI DAG visualizer showing real-time task execution progress
- [ ] Display per-node status: pending / running / consensus-passed / failed / escalated
- [ ] Export task graph as SVG or JSON for post-mortem analysis

### Cost Governor
- [ ] Implement a global **token budget** manager with per-session and per-model limits
- [ ] Add model invocation priority queue: prefer cheaper models for low-risk sub-tasks
- [ ] Provide a dry-run mode that estimates cost before executing
- [ ] Expose cost metrics in the TUI dashboard

### Observability & LLMOps
- [ ] Add structured logging per agent turn (role, model used, input tokens, output tokens, score, latency)
- [ ] Export metrics to Prometheus / OpenTelemetry
- [ ] Build a replay mode: replay a recorded session to debug or fine-tune consensus parameters
- [ ] Implement A/B testing harness for comparing different consensus threshold configurations

### Advanced Agent Patterns
- [ ] Explore `ReAct` (Reasoning + Acting) loop for the Architect role
- [ ] Explore `Chain-of-Thought` prompting for the Validator during complex security reviews
- [ ] Investigate multi-Executor parallel code generation with a "tournament" selection (best-of-N)
- [ ] Add a dedicated **Researcher** role for pulling in external documentation and Stack Overflow context via MCP web tools

---

## Backlog / Ideas

- [ ] Plugin system for adding custom roles and model adapters
- [ ] Support for non-Go codebases (Python, TypeScript) via language-agnostic AST tooling
- [ ] GitHub App integration for triggering consensus reviews on PRs
- [ ] Fine-tuned small model as a cheap "pre-screener" before invoking expensive frontier models
