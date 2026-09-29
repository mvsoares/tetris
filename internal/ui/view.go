package ui

import (
	"fmt"
	"strings"

	"tetris/internal/engine"
	"tetris/internal/i18n"

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

	// Every piece is rendered with a consistent height of exactly 2 lines.
	// For PieceI: row 0 is blank (8 spaces), row 1 has the 4 blocks ("████████").
	startRow := 0
	endRow := 2

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
	tr := i18n.Get(m.Language())
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
		overlay := PausedStyle.Render(tr.PausedTitle + "\n\n" + tr.PausedPrompt)
		return BoardBoxStyle.Render(lipgloss.Place(20, 20, lipgloss.Center, lipgloss.Center, overlay))
	} else if m.game.State == engine.StateGameOver {
		overlay := GameOverStyle.Render(
			fmt.Sprintf("%s\n\n%s\n%s\n%s\n\n%s\n%s",
				tr.GameOverTitle,
				tr.FormatGameOverScore(m.game.Score),
				tr.FormatGameOverLines(m.game.Lines),
				tr.FormatGameOverTetris(m.game.Tetrises),
				tr.GameOverRestart,
				tr.GameOverQuit,
			),
		)
		return BoardBoxStyle.Render(lipgloss.Place(20, 20, lipgloss.Center, lipgloss.Center, overlay))
	}

	return BoardBoxStyle.Render(boardContent)
}

func (m *Model) renderLeftPanel() string {
	tr := i18n.Get(m.Language())
	// 1. Hold Box (fixed height)
	holdContent := "        \n        "
	if m.game.HoldPiece != nil {
		holdContent = renderMiniPiece(m.game.HoldPiece.Type)
	}
	holdBox := PanelBoxStyle.Width(16).Height(4).Render(
		fmt.Sprintf("%s\n\n%s",
			HeaderLabelStyle.Render(tr.HoldHeader),
			lipgloss.NewStyle().Align(lipgloss.Center).Render(holdContent),
		),
	)

	// 2. Stats Box
	actionLine := " "
	if m.game.LastAction != "" {
		actionLine = ActionAlertStyle.Render(m.game.LastAction)
	}
	if m.game.UseLearned && m.game.LearnedModel != nil {
		probability := "—"
		if m.game.CurrentAIMove != nil && m.game.CurrentAIMove.MoveProbability != nil {
			probability = fmt.Sprintf("%.1f%%", *m.game.CurrentAIMove.MoveProbability*100)
		}
		actionLine += "\n" + tr.FormatSimHorizon(m.game.LearnedModel.Horizon, probability)
	}

	statsContent := fmt.Sprintf(
		"%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n%s",
		HeaderLabelStyle.Render(tr.Score),
		ValueStyle.Render(fmt.Sprintf("%d", m.game.Score)),
		HeaderLabelStyle.Render(tr.HighScore),
		ValueStyle.Render(fmt.Sprintf("%d", m.game.HighScore)),
		HeaderLabelStyle.Render(tr.Level),
		ValueStyle.Render(fmt.Sprintf("%d", m.game.Level)),
		HeaderLabelStyle.Render(tr.Lines),
		ValueStyle.Render(fmt.Sprintf("%d", m.game.Lines)),
		HeaderLabelStyle.Render(tr.Tetris),
		ValueStyle.Render(fmt.Sprintf("%d", m.game.Tetrises)),
		actionLine,
	)

	statsBox := PanelBoxStyle.Width(16).Height(15).Render(statsContent)

	return lipgloss.JoinVertical(lipgloss.Left, holdBox, statsBox)
}

func (m *Model) renderRightPanel() string {
	tr := i18n.Get(m.Language())
	// 1. Next Box (shows upcoming 2 pieces with fixed height)
	var nextPreviews []string
	for i := 0; i < 2; i++ {
		if i < len(m.game.NextQueue) {
			nextPreviews = append(nextPreviews, renderMiniPiece(m.game.NextQueue[i]))
		} else {
			nextPreviews = append(nextPreviews, "        \n        ")
		}
	}
	nextContent := strings.Join(nextPreviews, "\n\n")
	nextHeight := 7
	if m.game.LookaheadDepth > 0 {
		queue := make([]string, len(m.game.NextQueue))
		for i, piece := range m.game.NextQueue {
			queue[i] = string(piece)
		}
		nextContent += "\n\n" + strings.Join(queue, " ")
		nextHeight = 9
	}

	nextBox := PanelBoxStyle.Width(22).Height(nextHeight).Render(
		fmt.Sprintf("%s\n\n%s",
			HeaderLabelStyle.Render(tr.NextHeader),
			lipgloss.NewStyle().Align(lipgloss.Center).Render(nextContent),
		),
	)

	// Space key display label per language
	spaceKeyLabel := "Espaço "
	switch m.Language() {
	case i18n.LangEN, i18n.LangJA:
		spaceKeyLabel = "Space  "
	case i18n.LangFR:
		spaceKeyLabel = "Espace "
	case i18n.LangES:
		spaceKeyLabel = "Espacio"
	case i18n.LangIT:
		spaceKeyLabel = "Spazio "
	case i18n.LangDE:
		spaceKeyLabel = "Leert. "
	case i18n.LangRU:
		spaceKeyLabel = "Пробел "
	case i18n.LangZH:
		spaceKeyLabel = "空格键 "
	}

	// 2. Controls Box (fixed height 16 to comfortably fit all 14 rows)
	controlsContent := fmt.Sprintf(
		"%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s",
		KeyStyle.Render("← / →  "), DescStyle.Render(tr.KeyMove),
		KeyStyle.Render("↓      "), DescStyle.Render(tr.KeySoftDrop),
		KeyStyle.Render(spaceKeyLabel), DescStyle.Render(tr.KeyHardDrop),
		KeyStyle.Render("↑ / X  "), DescStyle.Render(tr.KeyRotateCW),
		KeyStyle.Render("Z      "), DescStyle.Render(tr.KeyRotateCCW),
		KeyStyle.Render("C / H  "), DescStyle.Render(tr.KeyHold),
		KeyStyle.Render("B / Tab"), DescStyle.Render(tr.KeyAutoPlay),
		KeyStyle.Render("M      "), DescStyle.Render(tr.KeySelectAI),
		KeyStyle.Render("L      "), DescStyle.Render(tr.KeySelectLang),
		KeyStyle.Render("4 / I  "), DescStyle.Render(tr.KeyFourLines),
		KeyStyle.Render("T      "), DescStyle.Render(tr.KeySetupTetris),
		KeyStyle.Render("P      "), DescStyle.Render(tr.KeyPause),
		KeyStyle.Render("R      "), DescStyle.Render(tr.KeyRestart),
		KeyStyle.Render("Q / Esc"), DescStyle.Render(tr.KeyQuit),
	)

	controlsBox := PanelBoxStyle.Width(22).Height(16).Render(
		fmt.Sprintf("%s\n\n%s",
			HeaderLabelStyle.Render(tr.ControlsHeader),
			controlsContent,
		),
	)

	return lipgloss.JoinVertical(lipgloss.Left, nextBox, controlsBox)
}

func (m *Model) renderLangMenu() string {
	tr := i18n.Get(m.Language())
	var options []string
	for i, langInfo := range i18n.AvailableLanguages {
		prefix := "  "
		if i == m.langMenuChoice {
			prefix = "> "
		}
		activeMark := ""
		if langInfo.Code == m.Language() {
			activeMark = "  ✓"
		}
		options = append(options, fmt.Sprintf("%s%d. %s %s (%s)%s", prefix, i+1, langInfo.Flag, langInfo.NativeName, langInfo.Name, activeMark))
	}
	content := HeaderLabelStyle.Render(tr.LangMenuTitle) + "\n\n" +
		strings.Join(options, "\n") + "\n\n" +
		tr.LangMenuHelp
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, PanelBoxStyle.Padding(1, 2).Render(content))
}

// View implements tea.Model View method.
func (m *Model) View() string {
	// Check terminal dimensions
	if m.width < MinTerminalWidth || m.height < MinTerminalHeight {
		return RenderTooSmallViewWithLang(m.width, m.height, m.Language())
	}

	tr := i18n.Get(m.Language())

	// Language selection modal
	if m.langMenuOpen {
		return m.renderLangMenu()
	}

	// AI selection modal
	if m.aiMenuOpen {
		options := []string{tr.AIOptionCurrent, tr.AIOptionLookahead, tr.AIOptionHybrid}
		if m.game.LearnedModel == nil {
			options[2] += tr.AIUnavailable
		}
		for i := range options {
			prefix := "  "
			if i == m.aiMenuChoice {
				prefix = "> "
			}
			options[i] = prefix + options[i]
		}
		content := HeaderLabelStyle.Render(tr.AIMenuTitle) + "\n\n" + strings.Join(options, "\n\n") +
			"\n\n" + tr.AIMenuHelp
		if m.game.LearnedModel != nil {
			content += "\n" + tr.FormatAIModelHorizon(m.game.LearnedModel.Horizon)
		}
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, PanelBoxStyle.Padding(1, 2).Render(content))
	}

	headerText := TitleStyle.Render(tr.Title)
	if m.game.AutoPlay {
		badgeColor := "#a6e3a1" // Green
		badgeText := tr.AutoPlayOn
		if engine.IsCleanupMode(m.game) {
			badgeColor = "#fab387" // Orange
			badgeText = tr.FormatAutoPlayCleanup(engine.CleanupThresholdPercent(m.game))
		} else if engine.IsHighSpeedMode(m.game) {
			badgeText = tr.AutoPlaySpeed
		}
		if m.game.LookaheadDepth > 0 {
			badgeText += tr.FormatLookahead(m.game.LookaheadDepth)
		}
		if m.game.UseLearned {
			badgeText += tr.HybridBadge
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
