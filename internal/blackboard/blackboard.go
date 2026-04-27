// Package blackboard provides the shared state layer that sits above MCP.
// Agents communicate via the Blackboard; tool execution is handled separately
// via MCP. Keeping them decoupled prevents race conditions on the filesystem.
//
// Phase 0: in-memory sync.Map.
// Phase 1: SQLite persistence + pub/sub subscription model.
package blackboard

import "sync"

// Blackboard is a thread-safe in-memory key-value store used as the
// communication bus between agents in the MCDD pipeline.
type Blackboard struct {
	store sync.Map
}

// New creates a new Blackboard instance.
func New() *Blackboard {
	return &Blackboard{}
}

// Set stores value under key.
func (b *Blackboard) Set(key string, value any) {
	b.store.Store(key, value)
}

// Get retrieves a value by key. Returns (value, true) if found.
func (b *Blackboard) Get(key string) (any, bool) {
	return b.store.Load(key)
}

// Delete removes a key.
func (b *Blackboard) Delete(key string) {
	b.store.Delete(key)
}

// Keys returns all keys currently in the board.
func (b *Blackboard) Keys() []string {
	var keys []string
	b.store.Range(func(k, _ any) bool {
		if s, ok := k.(string); ok {
			keys = append(keys, s)
		}
		return true
	})
	return keys
}
