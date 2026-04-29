package blackboard

import (
	"context"
	"sync"
)

// Blackboard defines the interface for the shared state layer between agents.
type Blackboard interface {
	// ── Session & Turn Management
	NewSession(id, path string) error
	GetLatestSession(path string) (*Session, error)
	CreateTurn(sessionID, prompt string, index int) (int64, error)
	
	// ── Task & Artifact Management
	UpsertTask(task *Task) error
	SaveArtifact(art *Artifact) error
	GetArtifactsForTurn(turnID int64) ([]*Artifact, error)

	// ── KV Persistence (Legacy Phase 0 support)
	Set(key string, value any)
	Get(key string) (any, bool)
	Delete(key string)
	Keys() []string

	// ── Pub/Sub
	Subscribe(ctx context.Context) <-chan any

	// ── Lifecycle
	Close() error
}

// MemoryBlackboard is the Phase 0 in-memory implementation.
type MemoryBlackboard struct {
	store sync.Map
}

func New() Blackboard { return &MemoryBlackboard{} }

func (b *MemoryBlackboard) Set(key string, value any)    { b.store.Store(key, value) }
func (b *MemoryBlackboard) Get(key string) (any, bool)  { return b.store.Load(key) }
func (b *MemoryBlackboard) Delete(key string)           { b.store.Delete(key) }
func (b *MemoryBlackboard) NewSession(id, path string) error { return nil }
func (b *MemoryBlackboard) GetLatestSession(path string) (*Session, error) { return nil, nil }
func (b *MemoryBlackboard) CreateTurn(s, p string, i int) (int64, error) { return 0, nil }
func (b *MemoryBlackboard) UpsertTask(t *Task) error    { return nil }
func (b *MemoryBlackboard) SaveArtifact(a *Artifact) error { return nil }
func (b *MemoryBlackboard) GetArtifactsForTurn(id int64) ([]*Artifact, error) { return nil, nil }
func (b *MemoryBlackboard) Subscribe(ctx context.Context) <-chan any { return nil }
func (b *MemoryBlackboard) Close() error                { return nil }
func (b *MemoryBlackboard) Keys() []string              { return nil }
