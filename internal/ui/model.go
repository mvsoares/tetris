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
	return tea.Tick(engine.AIActionInterval, func(t time.Time) tea.Msg {
		return aiTickMsg(t)
	})
}

type Model struct {
	game          *engine.Game
	width         int
	height        int
	quitting      bool
	nextGravityAt time.Time
	aiMenuOpen    bool
	aiMenuChoice  int
}

func NewModel() *Model {
	return NewModelWithLookahead(0, 4)
}

func NewModelWithLookahead(depth, width int) *Model {
	return NewModelWithLearned(depth, width, nil, false)
}

func NewModelWithLearned(depth, width int, model *engine.MoveModel, learned bool) *Model {
	g := engine.NewGame()
	g.ConfigureLookahead(depth, width)
	g.LearnedModel, g.UseLearned = model, learned
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
	m.nextGravityAt = time.Now().Add(m.game.TickInterval())
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
		if !m.aiMenuOpen && m.game.State == engine.StatePlaying {
			m.game.Tick()
		}
		m.nextGravityAt = time.Now().Add(m.game.TickInterval())
		return m, tickCmd(m.game.TickInterval())

	case aiTickMsg:
		if !m.aiMenuOpen && m.game.AutoPlay && m.game.State == engine.StatePlaying {
			remaining := m.game.TickInterval()
			if !m.nextGravityAt.IsZero() {
				remaining = time.Until(m.nextGravityAt)
			}
			m.game.SetAIGravityRemaining(remaining)
			m.game.StepAI()
		}
		return m, aiTickCmd()

	case tea.KeyMsg:
		key := msg.String()
		if m.aiMenuOpen && key != "q" && key != "ctrl+c" {
			switch key {
			case "esc", "m", "M":
				m.aiMenuOpen = false
			case "up", "down", "left", "right", "tab":
				if key == "up" || key == "left" {
					m.aiMenuChoice = (m.aiMenuChoice + 2) % 3
				} else {
					m.aiMenuChoice = (m.aiMenuChoice + 1) % 3
				}
			case "1":
				m.aiMenuChoice = 0
			case "2":
				m.aiMenuChoice = 1
			case "3":
				m.aiMenuChoice = 2
			case "enter", " ":
				depth := 0
				if m.aiMenuChoice == 1 {
					depth = 10
				}
				learned := m.aiMenuChoice == 2
				if learned && m.game.LearnedModel == nil {
					m.game.LastAction = "MODELO INDISPONÍVEL"
					return m, nil
				}
				if m.game.UseLearned != learned || m.game.LookaheadDepth != depth || (depth > 0 && m.game.BeamWidth != 4) {
					wasPaused := m.game.State == engine.StatePaused
					// Close the old session before changing policy metadata.
					m.game.Restart()
					m.game.ConfigureLookahead(depth, 4)
					m.game.UseLearned = learned
					if wasPaused {
						m.game.State = engine.StatePaused
					}
				}
				m.aiMenuOpen = false
			}
			return m, nil
		}
		switch msg.String() {
		case "m", "M":
			m.aiMenuOpen = true
			m.aiMenuChoice = 0
			if m.game.LookaheadDepth > 0 {
				m.aiMenuChoice = 1
			}
			if m.game.UseLearned {
				m.aiMenuChoice = 2
			}
			return m, nil
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
