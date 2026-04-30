// Package cli defines the cobra command tree for the `ca` CLI.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/agent/roles"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/blackboard"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/consensus"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/dag"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/sandbox"
	"github.com/objectisnotdefined/consensus-agent/ca/internal/tui"
	"github.com/objectisnotdefined/consensus-agent/ca/pkg/config"
	"github.com/objectisnotdefined/consensus-agent/ca/pkg/llm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "ca [workspace]",
	Short: "⚡ consensus-agent — MCDD powered AI engineering framework",
	Long: `consensus-agent orchestrates multiple AI models into a collaborative
expert team (Navigator, Architect, Executor, Validator), delivering a fully
automated pipeline from requirements analysis to code merge.

Powered by MCDD — Model Consensus Driven Development.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve workspace path
		workspace := "."
		if len(args) > 0 {
			workspace = args[0]
		}
		abs, err := filepath.Abs(workspace)
		if err != nil {
			return fmt.Errorf("invalid workspace path: %w", err)
		}
		if _, err := os.Stat(abs); err != nil {
			return fmt.Errorf("workspace not found: %s", abs)
		}

		// Build dependencies
		cfg := config.Default()
		// Load actual config from viper (if file exists)
		if err := viper.Unmarshal(&cfg); err != nil {
			// fallback to default on error
		}

		home, _ := os.UserHomeDir()
		dbPath := filepath.Join(home, ".config", "ca", "history.db")
		bb, err := blackboard.NewSQLiteBlackboard(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize blackboard: %w", err)
		}
		defer bb.Close()

		sessionID := uuid.New().String()
		if err := bb.NewSession(sessionID, abs); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		// Initialize LLM selector
		selector := llm.NewSelector(cfg.Models)

		// Initialize Sandbox
		sbManager := sandbox.NewManager()
		sbPath, cleanup, err := sbManager.Prepare(abs)
		if err != nil {
			return fmt.Errorf("failed to prepare sandbox: %w", err)
		}
		defer cleanup()
		bb.Set(consensus.KeySandboxPath, sbPath)

		registry := agent.NewRegistry([]agent.Agent{
			roles.NewNavigator(bb, selector),
			roles.NewArchitect(bb, selector),
			roles.NewExecutor(bb, selector),
			roles.NewValidator(bb, selector),
		})

		// Build the MCDD execution pipeline (DAG)
		pipeline, err := dag.MCDDPipeline()
		if err != nil {
			return fmt.Errorf("failed to build dag pipeline: %w", err)
		}
		dagExec := dag.NewExecutor(pipeline)

		// Initialize Consensus Evaluator
		evaluator := consensus.NewEvaluator(bb, cfg.Consensus.Threshold, cfg.Consensus.MaxRounds)

		// Launch TUI
		model := tui.New(registry, dagExec, evaluator, abs, cfg, bb, sessionID)
		p := tea.NewProgram(model,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}
		return nil
	},
}

// Execute runs the root command. Called by main.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "",
		"config file (default: ~/.config/ca/ca.yaml)",
	)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, _ := os.UserHomeDir()
		viper.AddConfigPath(filepath.Join(home, ".config", "ca"))
		viper.SetConfigName("ca")
		viper.SetConfigType("yaml")
	}
	viper.AutomaticEnv()
	_ = viper.ReadInConfig() // non-fatal: defaults are used if no file found
}
