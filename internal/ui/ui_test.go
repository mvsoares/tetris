package ui

import (
	"strings"
	"testing"
	"time"

	"tetris/internal/engine"

	tea "github.com/charmbracelet/bubbletea"
)

func menuKey(m *Model, key tea.KeyMsg) { m.Update(key) }

func runeKey(r rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }

func TestAIMenuFreezesAndCancelsWithoutChangingGame(t *testing.T) {
	m := NewModelWithLogger(nil)
	m.width, m.height = 100, 40
	m.game.AutoPlay = true
	m.game.CurrentAIMove = engine.FindBestMove(m.game)
	plan, piece, session := m.game.CurrentAIMove, *m.game.CurrentPiece, m.game.SessionID
	menuKey(m, runeKey('m'))
	if !m.aiMenuOpen || !strings.Contains(m.View(), "Lookahead 10 peças") {
		t.Fatal("missing AI menu")
	}
	m.Update(tickMsg(time.Now()))
	m.Update(aiTickMsg(time.Now()))
	for _, key := range []rune{'b', 'p', 'r', 'c'} {
		menuKey(m, runeKey(key))
	}
	if *m.game.CurrentPiece != piece || m.game.MoveCount != 0 || m.game.State != engine.StatePlaying || !m.game.AutoPlay || m.game.CurrentAIMove != plan {
		t.Fatal("game changed while menu open")
	}
	menuKey(m, tea.KeyMsg{Type: tea.KeyDown})
	menuKey(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.aiMenuOpen || m.quitting || m.game.LookaheadDepth != 0 || m.game.SessionID != session {
		t.Fatal("cancel changed policy or quit")
	}
	m.Update(tickMsg(time.Now()))
	if m.game.CurrentPiece.Y != piece.Y+1 {
		t.Fatal("gravity did not resume")
	}
}

func TestAIMenuSelectsPolicyAndRestartsOnlyOnChange(t *testing.T) {
	m := NewModelWithLogger(nil)
	m.game.AutoPlay, m.game.HighScore, m.game.Score = true, 900, 123
	oldSession := m.game.SessionID
	menuKey(m, runeKey('M'))
	menuKey(m, runeKey('2'))
	menuKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.aiMenuOpen || m.game.LookaheadDepth != 10 || m.game.BeamWidth != 4 || len(m.game.NextQueue) != 10 || m.game.Score != 0 || m.game.SessionID == oldSession || m.game.HighScore != 900 || !m.game.AutoPlay {
		t.Fatal("lookahead selection did not reset round and preserve settings")
	}
	m.game.Score = 456
	lookaheadSession := m.game.SessionID
	menuKey(m, runeKey('m'))
	if m.aiMenuChoice != 1 {
		t.Fatal("menu did not select current policy")
	}
	menuKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.game.Score != 456 || m.game.SessionID != lookaheadSession {
		t.Fatal("same policy restarted game")
	}
	m.game.State = engine.StatePaused
	menuKey(m, runeKey('m'))
	menuKey(m, tea.KeyMsg{Type: tea.KeyUp})
	menuKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.game.LookaheadDepth != 0 || m.game.Score != 0 || m.game.State != engine.StatePaused {
		t.Fatal("standard policy selection or pause preservation failed")
	}
	m.game.Restart()
	if m.game.LookaheadDepth != 0 {
		t.Fatal("restart lost menu selection")
	}
}

func TestAIMenuQuitStillWorks(t *testing.T) {
	m := NewModelWithLogger(nil)
	menuKey(m, runeKey('m'))
	_, cmd := m.Update(runeKey('q'))
	if !m.quitting || cmd == nil {
		t.Fatal("quit blocked by menu")
	}
}

func TestAIMenuLearnedChoiceAndUnavailableModel(t *testing.T) {
	m := NewModelWithLogger(nil)
	session := m.game.SessionID
	menuKey(m, runeKey('m'))
	menuKey(m, runeKey('3'))
	menuKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	if !m.aiMenuOpen || m.game.UseLearned || m.game.SessionID != session {
		t.Fatal("unavailable model selected")
	}
	m.game.LearnedModel = &engine.MoveModel{ID: "test", Horizon: 50}
	menuKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.aiMenuOpen || !m.game.UseLearned || m.game.LookaheadDepth != 0 || m.game.SessionID == session {
		t.Fatal("learned menu selection failed")
	}
	menuKey(m, runeKey('m'))
	if m.aiMenuChoice != 2 {
		t.Fatal("learned selection not highlighted")
	}
	menuKey(m, runeKey('1'))
	menuKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.game.UseLearned {
		t.Fatal("standard policy did not disable learned mode")
	}
}

func TestTooSmallView(t *testing.T) {
	out := RenderTooSmallView(50, 20)
	if !strings.Contains(out, "TERMINAL MUITO PEQUENO") {
		t.Errorf("expected warning header in small view, got: %s", out)
	}
	if !strings.Contains(out, "50 × 20") {
		t.Errorf("expected current dimensions in small view, got: %s", out)
	}
}

func TestModelResizeAndView(t *testing.T) {
	m := NewModel()

	// Initial view with 0 dimensions should show warning
	viewSmall := m.View()
	if !strings.Contains(viewSmall, "TERMINAL MUITO PEQUENO") {
		t.Errorf("expected warning when model dimensions are 0")
	}

	// Resize to sufficient dimensions
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	model := updated.(*Model)

	if model.width != 80 || model.height != 30 {
		t.Errorf("expected dimensions 80x30, got %dx%d", model.width, model.height)
	}

	viewLarge := model.View()
	if !strings.Contains(viewLarge, "T E T R I S   G O") {
		t.Errorf("expected title in view, got: %s", viewLarge)
	}
	if !strings.Contains(viewLarge, "HOLD") || !strings.Contains(viewLarge, "NEXT") {
		t.Errorf("expected side panels in view")
	}
	if !strings.Contains(viewLarge, "TETRIS") {
		t.Errorf("expected TETRIS counter in view")
	}
}

func TestModelKeyInput(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 30

	// Test Pause toggle
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model := updated.(*Model)
	if model.game.State != engine.StatePaused {
		t.Errorf("expected game to be paused after pressing 'p'")
	}

	// Unpause
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model = updated.(*Model)
	if model.game.State != engine.StatePlaying {
		t.Errorf("expected game to resume after pressing 'p'")
	}

	// Test shortcut '4' for 4 line pieces
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	model = updated.(*Model)
	if model.game.CurrentPiece.Type != engine.PieceI {
		t.Errorf("expected current piece to be I after pressing '4', got %s", model.game.CurrentPiece.Type)
	}

	// Test shortcut 't' for setup
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	model = updated.(*Model)
	if model.game.LastAction != "SETUP 4 LINHAS PRONTO!" {
		t.Errorf("expected setup action after pressing 't', got %s", model.game.LastAction)
	}

	// Test Auto-Play toggle 'b'
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(*Model)
	if !model.game.AutoPlay {
		t.Errorf("expected AutoPlay to be true after pressing 'b'")
	}
	viewAuto := model.View()
	if !strings.Contains(viewAuto, "AUTO-PLAY ON") {
		t.Errorf("expected view to contain AUTO-PLAY ON badge")
	}

	// Toggle off
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(*Model)
	if model.game.AutoPlay {
		t.Errorf("expected AutoPlay to be false after second 'b'")
	}
}

func TestTetrisCounterInUI(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 30

	m.game.Tetrises = 5
	viewPlaying := m.View()
	if !strings.Contains(viewPlaying, "TETRIS") {
		t.Errorf("expected 'TETRIS' label in playing view")
	}
	if !strings.Contains(viewPlaying, "5") {
		t.Errorf("expected '5' in playing view")
	}

	// In GameOver state
	m.game.State = engine.StateGameOver
	viewGameOver := m.View()
	if !strings.Contains(viewGameOver, "Tetris: 5") {
		t.Errorf("expected 'Tetris: 5' in GameOver overlay, got:\n%s", viewGameOver)
	}

	// Restart resets count
	m.game.Restart()
	if m.game.Tetrises != 0 {
		t.Errorf("expected Tetrises to be 0 after restart, got %d", m.game.Tetrises)
	}
}

func TestFixedPanelLayoutStability(t *testing.T) {
	// 1. Verify renderMiniPiece outputs exactly 2 lines for EVERY tetromino
	allPieces := append(engine.AllPieces, engine.TetrominoType(""))
	for _, p := range allPieces {
		mini := renderMiniPiece(p)
		lines := strings.Split(mini, "\n")
		if len(lines) != 2 {
			t.Errorf("expected piece %s to render exactly 2 lines, got %d:\n%q", p, len(lines), mini)
		}
	}

	// 2. Verify right panel height remains 100% constant regardless of pieces in next queue
	m := NewModel()
	m.width = 80
	m.height = 30

	// Case A: Next queue has two I pieces
	m.game.NextQueue = []engine.TetrominoType{engine.PieceI, engine.PieceI}
	panelA := m.renderRightPanel()
	heightA := len(strings.Split(panelA, "\n"))

	// Case B: Next queue has two 2-line pieces (e.g. O and T)
	m.game.NextQueue = []engine.TetrominoType{engine.PieceO, engine.PieceT}
	panelB := m.renderRightPanel()
	heightB := len(strings.Split(panelB, "\n"))

	// Case C: Mixed pieces (e.g. S and I)
	m.game.NextQueue = []engine.TetrominoType{engine.PieceS, engine.PieceI}
	panelC := m.renderRightPanel()
	heightC := len(strings.Split(panelC, "\n"))

	if heightA != heightB || heightB != heightC {
		t.Errorf("right panel height fluctuated: 2xI=%d, 2xO/T=%d, mixed=%d", heightA, heightB, heightC)
	}

	// 3. Verify total View() height is stable
	m.game.NextQueue = []engine.TetrominoType{engine.PieceI, engine.PieceI}
	viewA := m.View()
	linesA := len(strings.Split(viewA, "\n"))

	m.game.NextQueue = []engine.TetrominoType{engine.PieceO, engine.PieceT}
	viewB := m.View()
	linesB := len(strings.Split(viewB, "\n"))

	if linesA != linesB {
		t.Errorf("view height changed when pieces had 2-lines: viewA=%d, viewB=%d", linesA, linesB)
	}
}
