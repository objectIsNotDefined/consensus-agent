package llm

import (
	"fmt"
	"strings"

	"github.com/objectisnotdefined/consensus-agent/ca/pkg/config"
)

// Selector manages a pool of LLM clients and selects them by role.
type Selector struct {
	clients map[string]LLMClient
	configs []config.ModelConfig
}

// NewSelector initializes a selector from the given model configurations.
func NewSelector(configs []config.ModelConfig) *Selector {
	s := &Selector{
		clients: make(map[string]LLMClient),
		configs: configs,
	}

	for _, cfg := range configs {
		client := NewClient(cfg.Name, cfg.Provider, cfg.APIKey, cfg.EndpointURL)
		s.clients[cfg.Name] = client
	}

	return s
}

// SelectByRole returns the first available client that supports the given role.
// In Phase 1, it simply matches the role string against the configured roles.
func (s *Selector) SelectByRole(role string) (LLMClient, error) {
	for _, cfg := range s.configs {
		for _, r := range cfg.Roles {
			if strings.EqualFold(r, role) {
				client, ok := s.clients[cfg.Name]
				if ok {
					return client, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no model configured for role: %s", role)
}

// GetAllRoles returns all clients grouped by their roles.
func (s *Selector) GetAllRoles() map[string][]LLMClient {
	res := make(map[string][]LLMClient)
	for _, cfg := range s.configs {
		client := s.clients[cfg.Name]
		for _, r := range cfg.Roles {
			res[r] = append(res[r], client)
		}
	}
	return res
}
