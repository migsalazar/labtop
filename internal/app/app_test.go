package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelRendersTruthfulPlaceholder(t *testing.T) {
	t.Parallel()

	view := NewModel().View()

	for _, expected := range []string{
		"LABTOP // CONSOLE",
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

	updated, command := NewModel().Update(tea.WindowSizeMsg{Width: 100, Height: 30})
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

			_, command := NewModel().Update(test.key)
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
