package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Panel styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff007f")).
			MarginBottom(1)

	SubTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00f0f0"))

	PanelBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#44475a")).
			Padding(0, 1)

	BoardBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#bd93f9")).
			Padding(0, 0)

	HeaderLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ffb86c"))

	ValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#50fa7b"))

	KeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f1fa8c"))

	DescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4"))

	ActionAlertStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ff5555")).
				Blink(true)

	GameOverStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff5555")).
			Background(lipgloss.Color("#282a36")).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#ff5555")).
			Padding(1, 2).
			Align(lipgloss.Center)

	PausedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f1fa8c")).
			Background(lipgloss.Color("#282a36")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#f1fa8c")).
			Padding(1, 2).
			Align(lipgloss.Center)

	EmptyCellSymbol = " ·"
	GhostCellSymbol = "░░"
	BlockCellSymbol = "██"

	EmptyCellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#2f343f"))
	GhostCellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#5a5266"))
)
