// Package config defines the consensus-agent configuration schema.
// Values are loaded via Viper with the precedence:
// explicit flags > environment variables > config file > defaults.
package config

// Config is the root configuration structure.
type Config struct {
	Consensus ConsensusConfig `mapstructure:"consensus" json:"consensus"`
	Cost      CostConfig      `mapstructure:"cost" json:"cost"`
	Models    []ModelConfig   `mapstructure:"models" json:"models"`
}

// ConsensusConfig controls the multi-model consensus protocol.
type ConsensusConfig struct {
	// Threshold is the minimum confidence score [0, 1] required for output acceptance.
	Threshold float64 `mapstructure:"threshold" json:"threshold"`
	// MaxRounds is the maximum number of debate iterations before escalation.
	MaxRounds int `mapstructure:"max_rounds" json:"max_rounds"`
}

// CostConfig controls token budget enforcement.
type CostConfig struct {
	// TokenBudget is the total token limit per session across all model calls.
	TokenBudget int `mapstructure:"token_budget" json:"token_budget"`
}

// ModelConfig describes a single LLM integration (Phase 1).
type ModelConfig struct {
	Name        string   `mapstructure:"name" json:"name"`
	Provider    string   `mapstructure:"provider" json:"provider"`
	APIKey      string   `mapstructure:"apikey" json:"api_key"`
	EndpointURL string   `mapstructure:"endpointurl" json:"endpoint_url"`
	Roles       []string `mapstructure:"roles" json:"roles"`
}

// Default returns a Config populated with sensible defaults.
func Default() *Config {
	return &Config{
		Consensus: ConsensusConfig{
			Threshold: 0.85,
			MaxRounds: 3,
		},
		Cost: CostConfig{
			TokenBudget: 10000,
		},
	}
}
