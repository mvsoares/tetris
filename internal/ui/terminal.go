package ui

import (
	"fmt"

	"tetris/internal/i18n"

	"github.com/charmbracelet/lipgloss"
)

const (
	MinTerminalWidth  = 64
	MinTerminalHeight = 26
)

// RenderTooSmallView displays a clear warning when the terminal dimensions are insufficient, using default pt-br.
func RenderTooSmallView(width, height int) string {
	return RenderTooSmallViewWithLang(width, height, i18n.LangPTBR)
}

// RenderTooSmallViewWithLang displays a localized warning when the terminal dimensions are insufficient.
func RenderTooSmallViewWithLang(width, height int, lang i18n.Language) string {
	tr := i18n.Get(lang)

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
		"%s\n\n%s\n%s\n%s\n\n%s",
		titleStyle.Render(tr.TerminalTooSmallTitle),
		dimStyle.Render(tr.TerminalRequirement),
		accentStyle.Render(fmt.Sprintf(tr.TerminalRequired, fmt.Sprintf("%d × %d", MinTerminalWidth, MinTerminalHeight))),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#ff7700")).Bold(true).Render(fmt.Sprintf(tr.TerminalCurrent, fmt.Sprintf("%d × %d", width, height))),
		dimStyle.Render(tr.TerminalResizePrompt),
	)

	renderedBox := boxStyle.Render(content)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, renderedBox)
}
