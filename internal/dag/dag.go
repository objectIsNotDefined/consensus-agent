// Package dag implements the Directed Acyclic Graph (DAG) engine for
// orchestrating the MCDD agent pipeline.
//
// Agents are represented as nodes with explicit dependency edges. The executor
// uses Kahn's algorithm to derive a topological order, then releases nodes for
// execution as their dependencies complete.
package dag

import (
	"fmt"

	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
)

// NodeID is the unique identifier for a node within a DAG.
type NodeID string

// Node represents a single task unit in the pipeline.
type Node struct {
	ID   NodeID
	Role agent.Role // Which agent handles this node
	Deps []NodeID   // IDs of nodes that must complete before this one starts
}

// DAG is the directed acyclic graph describing the full execution plan.
type DAG struct {
	nodes   map[NodeID]*Node
	roleMap map[agent.Role]NodeID // reverse lookup: role → node ID
	order   []NodeID             // topologically sorted execution order
}

// Build constructs a DAG from a slice of nodes and validates it.
// Returns an error if the graph contains a cycle or references an unknown dep.
func Build(nodes []Node) (*DAG, error) {
	g := &DAG{
		nodes:   make(map[NodeID]*Node, len(nodes)),
		roleMap: make(map[agent.Role]NodeID, len(nodes)),
	}

	for i := range nodes {
		n := &nodes[i]
		if _, dup := g.nodes[n.ID]; dup {
			return nil, fmt.Errorf("dag: duplicate node ID %q", n.ID)
		}
		g.nodes[n.ID] = n
		g.roleMap[n.Role] = n.ID
	}

	// Validate all dependencies exist
	for _, n := range g.nodes {
		for _, dep := range n.Deps {
			if _, ok := g.nodes[dep]; !ok {
				return nil, fmt.Errorf("dag: node %q references unknown dep %q", n.ID, dep)
			}
		}
	}

	order, err := topoSort(g.nodes)
	if err != nil {
		return nil, err
	}
	g.order = order
	return g, nil
}

// Order returns the topologically sorted node IDs (roots first).
func (g *DAG) Order() []NodeID { return g.order }

// NodeByRole returns the node for a given agent role, or nil if not found.
func (g *DAG) NodeByRole(role agent.Role) *Node {
	id, ok := g.roleMap[role]
	if !ok {
		return nil
	}
	return g.nodes[id]
}

// Deps returns the direct dependency roles for the given role.
func (g *DAG) Deps(role agent.Role) []agent.Role {
	n := g.NodeByRole(role)
	if n == nil {
		return nil
	}
	roles := make([]agent.Role, 0, len(n.Deps))
	for _, depID := range n.Deps {
		if dep, ok := g.nodes[depID]; ok {
			roles = append(roles, dep.Role)
		}
	}
	return roles
}

// ── Kahn's topological sort ──────────────────────────────────────────────────

func topoSort(nodes map[NodeID]*Node) ([]NodeID, error) {
	// Build in-degree map and adjacency list
	inDegree := make(map[NodeID]int, len(nodes))
	children := make(map[NodeID][]NodeID, len(nodes))

	for id := range nodes {
		inDegree[id] = 0
	}
	for id, n := range nodes {
		for _, dep := range n.Deps {
			children[dep] = append(children[dep], id)
			inDegree[id]++
		}
	}

	// Collect all zero-in-degree nodes (roots)
	queue := make([]NodeID, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	order := make([]NodeID, 0, len(nodes))
	for len(queue) > 0 {
		// Pop front
		cur := queue[0]
		queue = queue[1:]
		order = append(order, cur)

		for _, child := range children[cur] {
			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	if len(order) != len(nodes) {
		return nil, fmt.Errorf("dag: cycle detected — graph is not a DAG")
	}
	return order, nil
}
