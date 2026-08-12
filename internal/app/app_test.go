package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/migsalazar/labtop/internal/config"
)

func TestTerminalColorFallbacks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		color     lipgloss.CompleteColor
		trueColor string
		ansi256   string
		ansi      string
	}{
		{name: "identity", color: identityColor, trueColor: "#9B7CFF", ansi256: "13", ansi: "13"},
		{name: "primary", color: primaryColor, trueColor: "#F0F0E6", ansi256: "15", ansi: "15"},
		{name: "secondary", color: secondaryColor, trueColor: "#7D828A", ansi256: "8", ansi: "8"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if test.color.TrueColor != test.trueColor || test.color.ANSI256 != test.ansi256 || test.color.ANSI != test.ansi {
				t.Fatalf("color = %#v, want TrueColor %q, ANSI256 %q, ANSI %q", test.color, test.trueColor, test.ansi256, test.ansi)
			}
		})
	}
}

func TestModelRendersTruthfulConfiguredPlaceholder(t *testing.T) {
	t.Parallel()

	configuration := config.Config{Console: config.Console{Title: "TEST // CONSOLE"}}
	view := NewModel(configuration).View()

	for _, expected := range []string{
		"TEST // CONSOLE",
		"Monitoring features are not implemented yet.",
		"q quit",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("View() = %q, want %q", view, expected)
		}
	}
}

func TestModelRecordsTerminalSize(t *testing.T) {
	t.Parallel()

	updated, command := NewModel(config.Config{}).Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if command != nil {
		t.Fatal("resize returned an unexpected command")
	}

	model, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update() model type = %T, want app.Model", updated)
	}
	if model.width != 100 || model.height != 30 {
		t.Fatalf("terminal size = %dx%d, want 100x30", model.width, model.height)
	}
}

func TestModelQuitKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{name: "q", key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{name: "control-c", key: tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, command := NewModel(config.Config{}).Update(test.key)
			if command == nil {
				t.Fatal("quit key returned no command")
			}
			message := command()
			if _, ok := message.(tea.QuitMsg); !ok {
				t.Fatalf("quit command message type = %T, want tea.QuitMsg", message)
			}
		})
	}
}
