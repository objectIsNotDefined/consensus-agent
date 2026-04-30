package sandbox

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

// Manager handles the creation and maintenance of isolated virtual workspaces.
type Manager struct {
	rootDir string
}

// NewManager creates a sandbox manager that stores temporary workspaces
// under the system's temp directory or a custom path.
func NewManager() *Manager {
	return &Manager{}
}

// Prepare creates a fresh copy of the target workspace in a temporary directory.
// It returns the path to the sandbox and a cleanup function.
func (m *Manager) Prepare(originalPath string) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "ca-sandbox-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// Copy files, excluding common heavy or sensitive directories
	err = m.copyDir(originalPath, tempDir)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to mirror workspace: %w", err)
	}

	return tempDir, cleanup, nil
}

// Diff generates a unified diff between the original workspace and the sandbox.
// It relies on the 'diff' command-line tool which is standard on Unix systems.
func (m *Manager) Diff(originalPath, sandboxPath string) (string, error) {
	// We use 'git diff --no-index' if available for better formatting,
	// otherwise fall back to standard 'diff -urN'.
	args := []string{"diff", "--no-index", "--color=never", originalPath, sandboxPath}
	cmd := exec.Command("git", args...)
	
	// Note: git diff --no-index returns exit code 1 if differences are found.
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return string(out), nil
		}
		// Fallback to standard diff
		cmd = exec.Command("diff", "-urN", originalPath, sandboxPath)
		out, _ = cmd.CombinedOutput()
		return string(out), nil
	}

	return string(out), nil
}

// copyDir recursively copies a directory, skipping ignored patterns.
func (m *Manager) copyDir(src, dst string) error {
	ignore := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		"bin":          true,
		"dist":         true,
		".gemini":      true,
	}

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip ignored directories
		if d.IsDir() && ignore[d.Name()] {
			return filepath.SkipDir
		}

		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		// Copy file
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}
