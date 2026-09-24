package ui

import (
	"time"

	"tetris/internal/engine"
	"tetris/internal/logger"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type aiTickMsg time.Time

func aiTickCmd() tea.Cmd {
	return tea.Tick(55*time.Millisecond, func(t time.Time) tea.Msg {
		return aiTickMsg(t)
	})
}

type Model struct {
	game     *engine.Game
	width    int
	height   int
	quitting bool
}

func NewModel() *Model {
	g := engine.NewGame()
	if playLogger, err := logger.NewLogger("logs/plays.jsonl"); err == nil {
		g.SetLogger(playLogger)
	}
	return &Model{
		game: g,
	}
}

func NewModelWithLogger(l *logger.Logger) *Model {
	g := engine.NewGame()
	g.SetLogger(l)
	return &Model{
		game: g,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(m.game.TickInterval()),
		aiTickCmd(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		if m.game.State == engine.StatePlaying {
			m.game.Tick()
		}
		return m, tickCmd(m.game.TickInterval())

	case aiTickMsg:
		if m.game.AutoPlay && m.game.State == engine.StatePlaying {
			m.game.StepAI()
		}
		return m, aiTickCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			_ = m.game.CloseLogger()
			return m, tea.Quit

		case "r", "R":
			m.game.Restart()
			return m, nil

		case "p", "P":
			m.game.TogglePause()
			return m, nil

		case "b", "B", "tab":
			m.game.ToggleAutoPlay()
			return m, nil
		}

		if m.game.State == engine.StatePlaying {
			switch msg.String() {
			case "left", "a", "h":
				m.game.MoveLeft()
			case "right", "d", "l":
				m.game.MoveRight()
			case "down", "s", "j":
				m.game.SoftDrop()
			case "up", "w", "k", "x":
				m.game.RotateCW()
			case "z":
				m.game.RotateCCW()
			case " ":
				m.game.HardDrop()
			case "c", "C":
				m.game.Hold()
			case "4", "i", "I":
				m.game.QueueLinePieces(4)
			case "t", "T":
				m.game.SetupFourLines()
			}
		}
	}

	return m, nil
}
