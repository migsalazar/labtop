package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/migsalazar/labtop/internal/config"
)

var (
	identityColor = lipgloss.CompleteColor{
		TrueColor: "#9B7CFF",
		ANSI256:   "13",
		ANSI:      "13",
	}
	primaryColor = lipgloss.CompleteColor{
		TrueColor: "#F0F0E6",
		ANSI256:   "15",
		ANSI:      "15",
	}
	secondaryColor = lipgloss.CompleteColor{
		TrueColor: "#7D828A",
		ANSI256:   "8",
		ANSI:      "8",
	}

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(identityColor)
	bodyStyle = lipgloss.NewStyle().
			Foreground(primaryColor)
	hintStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)
)

// Model is the initial Labtop Bubble Tea model.
type Model struct {
	config config.Config
	width  int
	height int
}

// NewModel returns the truthful initial placeholder model.
func NewModel(configuration config.Config) Model {
	return Model{config: configuration}
}

// Init starts no background work in the current scaffold.
func (Model) Init() tea.Cmd {
	return nil
}

// Update handles terminal sizing and the two immediate quit keys.
func (model Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.KeyMsg:
		switch message.String() {
		case "q", "ctrl+c":
			return model, tea.Quit
		}
	case tea.WindowSizeMsg:
		model.width = message.Width
		model.height = message.Height
	}

	return model, nil
}

// View renders only truthful scaffold state.
func (model Model) View() string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(model.config.Console.Title),
		"",
		bodyStyle.Render("Monitoring features are not implemented yet."),
		"",
		hintStyle.Render("q quit"),
	)

	if model.width <= 0 || model.height <= 0 {
		return content
	}

	return lipgloss.Place(
		model.width,
		model.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// Run starts Labtop in the terminal alternate screen.
func Run(configuration config.Config) error {
	program := tea.NewProgram(NewModel(configuration), tea.WithAltScreen())
	_, err := program.Run()
	return err
}
