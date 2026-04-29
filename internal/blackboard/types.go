package blackboard

import (
	"time"

	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
)

// Session represents a persistent conversation about a specific workspace.
type Session struct {
	ID            string    `json:"id"`
	WorkspacePath string    `json:"workspace_path"`
	CreatedAt     time.Time `json:"created_at"`
}

// Turn represents a single round of user-agent interaction.
type Turn struct {
	ID             int64     `json:"id"`
	SessionID      string    `json:"session_id"`
	Index          int       `json:"index"` // 0-based
	UserPrompt     string    `json:"user_prompt"`
	SystemResponse string    `json:"system_response"`
	CreatedAt      time.Time `json:"created_at"`
}

// Task represents an atomic action planned or executed by an agent.
type Task struct {
	ID         string       `json:"id"`
	TurnID     int64        `json:"turn_id"`
	Role       agent.Role   `json:"role"`
	Status     agent.Status `json:"status"`
	InputData  string       `json:"input_data"`
	OutputData string       `json:"output_data"`
	CreatedAt  time.Time    `json:"created_at"`
}

// Artifact represents a versioned output like code or tests.
type Artifact struct {
	ID              string    `json:"id"`
	TaskID          string    `json:"task_id"`
	FilePath        string    `json:"file_path"`
	Content         string    `json:"content"`
	ConfidenceScore float64   `json:"confidence_score"`
	Version         int       `json:"version"`
	CreatedAt       time.Time `json:"created_at"`
}
