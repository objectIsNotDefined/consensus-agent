# Consensus-Agent (CA)

> **"Where Models Agree, Code Excels."**

`Consensus-Agent` is a high-performance AI software engineering framework built in **Golang**. Powered by the unique **MCDD (Model-Consensus Driven Development)** architecture, it orchestrates the world's leading AI models into a collaborative distributed expert team — delivering a fully automated pipeline from requirements analysis to code merge.

---

## 🌟 Vision

In the era of single-model AI, developers are constantly plagued by hallucinations, limited context windows, and logical inconsistencies in generated code.

**Consensus-Agent's vision is to break the "single-point intelligence" bottleneck.**

By employing a distributed consensus mechanism, we build an "AI R&D Brigade" capable of self-awareness, mutual auditing, and parallel execution. High-quality code should not emerge from a single model's guess — it should be the **consensus output** forged through deep deliberation between multiple specialized expert models.

---

## 🏗 Architecture

### 1. The Specialized Swarm

The project adopts a **1+N Role Matrix**. Instead of binding roles to specific model names (which age quickly), each role is defined by a **Capability Profile**. The framework dynamically selects the best available model at runtime based on declared requirements — similar to how Kubernetes schedules Pods.

| Role | Capability Profile | Responsibility |
| :--- | :--- | :--- |
| **Navigator** | `context_window > 1M tokens` | **Global awareness hub.** Maintains a real-time semantic index of the entire codebase and technical documentation. |
| **Architect** | `strong reasoning & planning` | **Logic and decision hub.** Handles task decomposition, architecture selection, and conflict arbitration between models. |
| **Executor** | `strong code generation` | **Agile coding engine.** Focuses on business logic implementation, high-performance refactoring, and complex function authoring. |
| **Validator** | `code review + SAST toolchain` | **Quality and security guardian.** Combines LLM semantic auditing with static analysis tools (e.g., `golangci-lint`, `gosec`) for dual-track verification. |

### 2. Consensus Protocol

`Consensus-Agent` is not a simple API wrapper. It is a rigorous collaboration loop:

- **State Awareness:** All roles share a common **Blackboard** — a persistent SQLite-backed shared state layer. This enables **multi-turn dialogue**, allowing users to refine code over several rounds while maintaining full context of previous changes.
- **Parallel Sprints:** Once a task begins, the Executor writes code logic while the Validator simultaneously generates test stubs — achieving "develop-and-test in parallel."
- **Logical Consensus:** Before any code is written to disk, the Architect cross-compares the Executor's output and the Validator's audit result. A **Confidence Score** is computed. Only code that meets the consensus threshold is merged.
- **Consensus Failure Handling:** If the confidence score falls below the threshold after N rounds of debate, the system triggers a **Human-in-the-Loop** checkpoint, pausing for developer review.

### 3. Circuit Breaker

Each model integration is protected by a **Circuit Breaker**. If a model exceeds error rate or latency thresholds, it is automatically replaced by a fallback model. This ensures the pipeline remains resilient against API timeouts, rate limits, or malformed outputs.

---

## 🚀 Key Features

- **Multi-Model Parallel Scheduling:** Leverages Golang's Goroutine model for true async concurrent inference across multiple models, significantly reducing end-to-end latency.
- **Capability-Based Model Selection:** Roles are defined by capability profiles, not model names. The framework selects the best available model dynamically, making the system future-proof and API-key agnostic.
- **Confidence Score System:** Every code output is scored against the consensus threshold. Low-confidence results trigger additional debate rounds or human escalation.
- **Blackboard Shared State:** A dedicated shared state layer decouples agent communication from tool execution, eliminating race conditions in concurrent multi-agent workflows.
- **Dual-Track Validation:** The Validator combines LLM-based semantic code review with real static analysis tool output for robust, non-hallucinated quality gates.
- **Virtual Workspace:** Code simulation and testing run inside an isolated virtual environment, keeping local production environments clean and safe.
- **Self-Evolving Skill Library:** Successful refactoring and fix paths are persisted to a Vector Database, enabling RAG-powered recall during future decisions. The system genuinely learns from its own history.
- **Hardcore CLI / TUI:** Deep terminal workflow integration with interactive TUI, code diff previews, and fully automated PR generation.
- **Cost Governor:** A token budget manager controls parallel API call costs, enforcing spending limits and prioritizing model invocations to prevent runaway expenses.

---

## 🛠 Tech Stack

| Layer | Technology |
| :--- | :--- |
| **Language** | Golang 1.24+ |
| **Agent Communication** | Model Context Protocol (MCP) |
| **Shared State (Blackboard)** | In-memory KV / SQLite |
| **Memory & RAG** | Vector DB (e.g., Chroma) |
| **Orchestration** | Directed Acyclic Graph (DAG) Executor |
| **Static Analysis** | `golangci-lint`, `gosec` |
| **TUI Framework** | Bubble Tea (Charm.sh) |
| **License** | Apache-2.0 |

---

## 💡 Philosophy

> *"Individual models process data; the Consensus builds systems."*

See [TODO.md](./TODO.md) for the full development roadmap.