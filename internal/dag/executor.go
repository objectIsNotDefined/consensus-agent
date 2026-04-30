package dag

import (
	"sync"

	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
)

// Executor tracks runtime completion state and determines which nodes are
// ready to run based on their dependency graph.
//
// It is safe for concurrent use: MarkDone may be called from multiple
// goroutines (e.g. different Bubble Tea message handlers).
type Executor struct {
	dag  *DAG
	done map[NodeID]bool
	mu   sync.Mutex
}

// NewExecutor creates an Executor backed by the given DAG.
func NewExecutor(g *DAG) *Executor {
	return &Executor{
		dag:  g,
		done: make(map[NodeID]bool, len(g.nodes)),
	}
}

// Ready returns the roles that are immediately runnable: all of their
// dependencies are done and they have not yet been started.
//
// Typically called once at the start of a session to find root nodes.
func (e *Executor) Ready() []agent.Role {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.ready()
}

// MarkDone records that the agent for the given role has finished and returns
// the roles that became newly unblocked as a result.
//
// Returns an empty slice when no new nodes are unblocked (e.g. still waiting
// on other dependencies).
func (e *Executor) MarkDone(role agent.Role) []agent.Role {
	e.mu.Lock()
	defer e.mu.Unlock()

	n := e.dag.NodeByRole(role)
	if n == nil {
		return nil
	}
	e.done[n.ID] = true

	// Collect children of this node whose all deps are now satisfied
	var unlocked []agent.Role
	for _, child := range e.children(n.ID) {
		childNode := e.dag.nodes[child]
		if e.done[child] {
			continue // already processed
		}
		if e.allDepsDone(childNode) {
			unlocked = append(unlocked, childNode.Role)
		}
	}
	return unlocked
}

// IsComplete returns true when every node in the DAG has been marked done.
func (e *Executor) IsComplete() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.done) == len(e.dag.nodes)
}

// Reset clears all completion state, allowing the executor to be reused for a
// new session with the same DAG.
func (e *Executor) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.done = make(map[NodeID]bool, len(e.dag.nodes))
}

// ResetRoles removes the given roles from the completion map, making them
// available for re-execution without resetting the entire pipeline.
// Used by the Debate Loop to replay only Executor and Validator.
func (e *Executor) ResetRoles(roles []agent.Role) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, role := range roles {
		n := e.dag.NodeByRole(role)
		if n != nil {
			delete(e.done, n.ID)
		}
	}
}

// ── internal helpers (must be called with mu held) ───────────────────────────

func (e *Executor) ready() []agent.Role {
	var roles []agent.Role
	for _, id := range e.dag.order {
		if e.done[id] {
			continue
		}
		n := e.dag.nodes[id]
		if e.allDepsDone(n) {
			roles = append(roles, n.Role)
		}
	}
	return roles
}

func (e *Executor) allDepsDone(n *Node) bool {
	for _, dep := range n.Deps {
		if !e.done[dep] {
			return false
		}
	}
	return true
}

// children returns the IDs of nodes that directly depend on parentID.
func (e *Executor) children(parentID NodeID) []NodeID {
	var out []NodeID
	for id, n := range e.dag.nodes {
		for _, dep := range n.Deps {
			if dep == parentID {
				out = append(out, id)
				break
			}
		}
	}
	return out
}
