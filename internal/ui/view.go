package ui

import (
	"fmt"
	"strings"

	"tetris/internal/engine"

	"github.com/charmbracelet/lipgloss"
)

func renderMiniPiece(t engine.TetrominoType) string {
	if t == "" {
		return "        \n        "
	}

	shape := engine.ShapeFor(t, 0)
	var b strings.Builder
	color := engine.PieceColors[t]
	blockStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))

	// Find non-empty rows to keep it compact
	startRow := 0
	endRow := len(shape)
	if t == engine.PieceI {
		startRow = 1
		endRow = 2
	} else if len(shape) == 3 {
		startRow = 0
		endRow = 2
	}

	for r := startRow; r < endRow; r++ {
		rowStr := ""
		// Horizontal padding for alignment
		padLeft := 0
		if t == engine.PieceO {
			padLeft = 2
		} else if len(shape) == 3 {
			padLeft = 1
		}
		rowStr += strings.Repeat(" ", padLeft)

		for c := 0; c < len(shape[r]); c++ {
			if shape[r][c] != 0 {
				rowStr += blockStyle.Render(BlockCellSymbol)
			} else {
				rowStr += "  "
			}
		}
		b.WriteString(rowStr)
		if r < endRow-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m *Model) renderBoard() string {
	b := m.game.Board
	currPiece := m.game.CurrentPiece

	// Map of active piece blocks
	pieceCoords := make(map[[2]int]bool)
	if currPiece != nil {
		for _, pt := range currPiece.BlockCoords() {
			pieceCoords[pt] = true
		}
	}

	// Map of ghost piece blocks
	ghostCoords := make(map[[2]int]bool)
	if currPiece != nil && m.game.State == engine.StatePlaying {
		ghostY := b.GetGhostY(currPiece)
		ghost := currPiece.Clone()
		ghost.Y = ghostY
		for _, pt := range ghost.BlockCoords() {
			if !pieceCoords[pt] {
				ghostCoords[pt] = true
			}
		}
	}

	var sb strings.Builder
	for y := 0; y < engine.BoardHeight; y++ {
		for x := 0; x < engine.BoardWidth; x++ {
			pt := [2]int{x, y}
			if pieceCoords[pt] {
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(currPiece.Color))
				sb.WriteString(style.Render(BlockCellSymbol))
			} else if ghostCoords[pt] {
				sb.WriteString(GhostCellStyle.Render(GhostCellSymbol))
			} else if b.Cells[y][x].Filled {
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(b.Cells[y][x].Color))
				sb.WriteString(style.Render(BlockCellSymbol))
			} else {
				sb.WriteString(EmptyCellStyle.Render(EmptyCellSymbol))
			}
		}
		if y < engine.BoardHeight-1 {
			sb.WriteString("\n")
		}
	}

	boardContent := sb.String()

	// If paused or game over, display overlay
	if m.game.State == engine.StatePaused {
		overlay := PausedStyle.Render("   PAUSADO   \n\nPressione P\npara continuar")
		return BoardBoxStyle.Render(lipgloss.Place(20, 20, lipgloss.Center, lipgloss.Center, overlay))
	} else if m.game.State == engine.StateGameOver {
		overlay := GameOverStyle.Render(
			fmt.Sprintf(" GAME OVER \n\nPontos: %d\nLinhas: %d\nTetris: %d\n\n[R] Reiniciar\n[Q] Sair",
				m.game.Score, m.game.Lines, m.game.Tetrises),
		)
		return BoardBoxStyle.Render(lipgloss.Place(20, 20, lipgloss.Center, lipgloss.Center, overlay))
	}

	return BoardBoxStyle.Render(boardContent)
}

func (m *Model) renderLeftPanel() string {
	// 1. Hold Box
	holdContent := "        \n        "
	if m.game.HoldPiece != nil {
		holdContent = renderMiniPiece(m.game.HoldPiece.Type)
	}
	holdBox := PanelBoxStyle.Width(16).Render(
		fmt.Sprintf("%s\n\n%s",
			HeaderLabelStyle.Render("HOLD [C]"),
			lipgloss.NewStyle().Align(lipgloss.Center).Render(holdContent),
		),
	)

	// 2. Stats Box
	actionLine := ""
	if m.game.LastAction != "" {
		actionLine = fmt.Sprintf("\n\n%s", ActionAlertStyle.Render(m.game.LastAction))
	}

	statsContent := fmt.Sprintf(
		"%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n%s\n%s%s",
		HeaderLabelStyle.Render("SCORE"),
		ValueStyle.Render(fmt.Sprintf("%d", m.game.Score)),
		HeaderLabelStyle.Render("HIGH SCORE"),
		ValueStyle.Render(fmt.Sprintf("%d", m.game.HighScore)),
		HeaderLabelStyle.Render("LEVEL"),
		ValueStyle.Render(fmt.Sprintf("%d", m.game.Level)),
		HeaderLabelStyle.Render("LINES"),
		ValueStyle.Render(fmt.Sprintf("%d", m.game.Lines)),
		HeaderLabelStyle.Render("TETRIS"),
		ValueStyle.Render(fmt.Sprintf("%d", m.game.Tetrises)),
		actionLine,
	)

	statsBox := PanelBoxStyle.Width(16).Render(statsContent)

	return lipgloss.JoinVertical(lipgloss.Left, holdBox, statsBox)
}

func (m *Model) renderRightPanel() string {
	// 1. Next Box (shows upcoming 2 pieces to balance vertical height)
	var nextPreviews []string
	for i := 0; i < len(m.game.NextQueue) && i < 2; i++ {
		nextPreviews = append(nextPreviews, renderMiniPiece(m.game.NextQueue[i]))
	}
	nextContent := strings.Join(nextPreviews, "\n\n")

	nextBox := PanelBoxStyle.Width(20).Render(
		fmt.Sprintf("%s\n\n%s",
			HeaderLabelStyle.Render("NEXT"),
			lipgloss.NewStyle().Align(lipgloss.Center).Render(nextContent),
		),
	)

	// 2. Controls Box
	controlsContent := fmt.Sprintf(
		"%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s",
		KeyStyle.Render("← / →  "), DescStyle.Render("Mover"),
		KeyStyle.Render("↓      "), DescStyle.Render("Soft Drop"),
		KeyStyle.Render("Espaço "), DescStyle.Render("Hard Drop"),
		KeyStyle.Render("↑ / X  "), DescStyle.Render("Girar Horário"),
		KeyStyle.Render("Z      "), DescStyle.Render("Girar Anti-h"),
		KeyStyle.Render("C / H  "), DescStyle.Render("Guardar Peça"),
		KeyStyle.Render("B / Tab"), DescStyle.Render("Auto-Play (IA)"),
		KeyStyle.Render("4 / I  "), DescStyle.Render("4 Linhas (I)"),
		KeyStyle.Render("T      "), DescStyle.Render("Setup Tetris"),
		KeyStyle.Render("P      "), DescStyle.Render("Pausar"),
		KeyStyle.Render("R      "), DescStyle.Render("Reiniciar"),
		KeyStyle.Render("Q / Esc"), DescStyle.Render("Sair"),
	)

	controlsBox := PanelBoxStyle.Width(20).Render(
		fmt.Sprintf("%s\n\n%s",
			HeaderLabelStyle.Render("CONTROLES"),
			controlsContent,
		),
	)

	return lipgloss.JoinVertical(lipgloss.Left, nextBox, controlsBox)
}

// View implements tea.Model View method.
func (m *Model) View() string {
	// Check terminal dimensions
	if m.width < MinTerminalWidth || m.height < MinTerminalHeight {
		return RenderTooSmallView(m.width, m.height)
	}

	headerText := TitleStyle.Render("🎮  T E T R I S   G O  🎮")
	if m.game.AutoPlay {
		badgeColor := "#a6e3a1" // Green
		badgeText := "🤖 AUTO-PLAY ON"
		if engine.IsCleanupMode(m.game) {
			badgeColor = "#fab387" // Orange
			badgeText = "🤖 AUTO-PLAY [🚨 LIMPEZA 65%+]"
		}
		autoBadge := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#11111b")).
			Background(lipgloss.Color(badgeColor)).
			Padding(0, 1).
			Render(badgeText)
		headerText = lipgloss.JoinHorizontal(lipgloss.Center, headerText, "  ", autoBadge)
	}

	leftPanel := m.renderLeftPanel()
	boardPanel := m.renderBoard()
	rightPanel := m.renderRightPanel()

	gameRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		"  ",
		boardPanel,
		"  ",
		rightPanel,
	)

	fullContent := lipgloss.JoinVertical(
		lipgloss.Center,
		headerText,
		gameRow,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, fullContent)
}
