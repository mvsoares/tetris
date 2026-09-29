package ui

import (
	"strings"
	"testing"

	"tetris/internal/i18n"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLanguageMenuToggleAndNavigation(t *testing.T) {
	m := NewModelWithLogger(nil)
	m.width, m.height = 100, 40

	if m.Language() != i18n.LangPTBR {
		t.Fatalf("expected default language to be %s, got %s", i18n.LangPTBR, m.Language())
	}

	// Press 'L' to open language menu
	menuKey(m, runeKey('L'))
	if !m.langMenuOpen {
		t.Fatal("expected language menu to be open")
	}

	view := m.View()
	if !strings.Contains(view, "ESCOLHER IDIOMA") {
		t.Fatalf("expected language menu title in view, got: %s", view)
	}
	if !strings.Contains(view, "English") || !strings.Contains(view, "Español") || !strings.Contains(view, "Français") || !strings.Contains(view, "Italiano") {
		t.Fatalf("expected all languages in menu view, got: %s", view)
	}

	// Navigate with arrows
	menuKey(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.langMenuChoice != 1 {
		t.Fatalf("expected choice 1 after Down, got %d", m.langMenuChoice)
	}

	// Press Esc to close without changing
	menuKey(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.langMenuOpen {
		t.Fatal("expected language menu to be closed after Esc")
	}
	if m.Language() != i18n.LangPTBR {
		t.Fatalf("expected language to remain PTBR, got %s", m.Language())
	}
}

func TestLanguageSelectionUpdatesUI(t *testing.T) {
	testCases := []struct {
		keyRune     rune
		expected    i18n.Language
		checkString string
	}{
		{'2', i18n.LangEN, "CONTROLS"},
		{'3', i18n.LangES, "PUNTOS"},
		{'4', i18n.LangFR, "COMMANDES"},
		{'5', i18n.LangIT, "CONTROLLI"},
		{'6', i18n.LangDE, "STEUERUNG"},
		{'7', i18n.LangRU, "УПРАВЛЕНИЕ"},
		{'8', i18n.LangJA, "操作方法"},
		{'9', i18n.LangZH, "操作说明"},
		{'1', i18n.LangPTBR, "CONTROLES"},
	}

	for _, tc := range testCases {
		m := NewModelWithLogger(nil)
		m.width, m.height = 100, 40

		// Open language menu
		menuKey(m, runeKey('l'))
		// Select language by number
		menuKey(m, runeKey(tc.keyRune))
		// Confirm with Enter
		menuKey(m, tea.KeyMsg{Type: tea.KeyEnter})

		if m.langMenuOpen {
			t.Fatalf("expected language menu to be closed after selection for %s", tc.expected)
		}
		if m.Language() != tc.expected {
			t.Fatalf("expected language %s, got %s", tc.expected, m.Language())
		}

		view := m.View()
		if !strings.Contains(view, tc.checkString) {
			t.Errorf("expected view for %s to contain %q, got:\n%s", tc.expected, tc.checkString, view)
		}
	}
}

func TestDirectLanguageSetting(t *testing.T) {
	m := NewModelWithLearnedAndLang(0, 4, nil, false, i18n.LangEN)
	m.width, m.height = 100, 40

	if m.Language() != i18n.LangEN {
		t.Fatalf("expected language EN, got %s", m.Language())
	}

	view := m.View()
	if !strings.Contains(view, "CONTROLS") {
		t.Fatalf("expected 'CONTROLS' in English view")
	}

	// Change to Italian
	m.SetLanguage(i18n.LangIT)
	if m.Language() != i18n.LangIT {
		t.Fatalf("expected language IT, got %s", m.Language())
	}

	viewIT := m.View()
	if !strings.Contains(viewIT, "CONTROLLI") {
		t.Fatalf("expected 'CONTROLLI' in Italian view")
	}
}
