package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

const (
	MinTerminalWidth  = 64
	MinTerminalHeight = 26
)

// RenderTooSmallView displays a clear warning when the terminal dimensions are insufficient.
func RenderTooSmallView(width, height int) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#ff5555")).
		Padding(1, 3).
		Align(lipgloss.Center)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ff5555"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#aaaaaa"))

	accentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00f0f0")).
		Bold(true)

	content := fmt.Sprintf(
		"%s\n\n%s\n  Exigido: %s\n  Atual:   %s\n\n%s",
		titleStyle.Render("⚠️  TERMINAL MUITO PEQUENO  ⚠️"),
		dimStyle.Render("Para uma exibição adequada do Tetris:"),
		accentStyle.Render(fmt.Sprintf("%d × %d", MinTerminalWidth, MinTerminalHeight)),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#ff7700")).Bold(true).Render(fmt.Sprintf("%d × %d", width, height)),
		dimStyle.Render("Por favor, aumente o tamanho da sua janela."),
	)

	renderedBox := boxStyle.Render(content)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, renderedBox)
}
