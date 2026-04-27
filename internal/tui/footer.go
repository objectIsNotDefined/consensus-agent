package tui

import (
	"strings"

	"github.com/objectisnotdefined/consensus-agent/ca/internal/tui/styles"
)

type keyBinding struct {
	key  string
	desc string
}

var inputModeBindings = []keyBinding{
	{"Enter", "Start agents"},
	{"Ctrl+C", "Quit"},
}

var runModeBindings = []keyBinding{
	{"Tab", "Switch agent"},
	{"↑↓", "Scroll log"},
	{"R", "New task"},
	{"Q", "Quit"},
}

// renderFooter renders the bottom keybinding hint bar.
func renderFooter(width int, inputMode bool) string {
	bindings := runModeBindings
	if inputMode {
		bindings = inputModeBindings
	}

	parts := make([]string, 0, len(bindings))
	for _, b := range bindings {
		hint := styles.KeyHint.Render("["+b.key+"]") + " " +
			styles.Muted.Render(b.desc)
		parts = append(parts, hint)
	}

	content := strings.Join(parts, styles.Muted.Render("  ·  "))
	return styles.FooterStyle.Width(width).Render(content)
}
