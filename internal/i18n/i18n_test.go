package i18n

import (
	"strings"
	"testing"
)

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected Language
	}{
		// English
		{"en", LangEN},
		{"EN", LangEN},
		{"english", LangEN},
		{"English", LangEN},
		{"ing", LangEN},
		{"ingles", LangEN},
		{"inglês", LangEN},

		// Spanish
		{"es", LangES},
		{"ES", LangES},
		{"spanish", LangES},
		{"Spanish", LangES},
		{"espanol", LangES},
		{"español", LangES},
		{"esp", LangES},

		// PT-BR
		{"pt", LangPTBR},
		{"pt-br", LangPTBR},
		{"PT-BR", LangPTBR},
		{"pt_br", LangPTBR},
		{"portugues", LangPTBR},
		{"português", LangPTBR},
		{"portuguese", LangPTBR},

		// French
		{"fr", LangFR},
		{"FR", LangFR},
		{"french", LangFR},
		{"French", LangFR},
		{"francais", LangFR},
		{"français", LangFR},
		{"fra", LangFR},

		// Italian
		{"it", LangIT},
		{"IT", LangIT},
		{"ita", LangIT},
		{"ITA", LangIT},
		{"italian", LangIT},
		{"Italian", LangIT},
		{"italiano", LangIT},

		// German
		{"de", LangDE},
		{"DE", LangDE},
		{"german", LangDE},
		{"deutsch", LangDE},
		{"ger", LangDE},
		{"alemao", LangDE},
		{"alemão", LangDE},

		// Russian
		{"ru", LangRU},
		{"RU", LangRU},
		{"russian", LangRU},
		{"rus", LangRU},
		{"russo", LangRU},

		// Japanese
		{"ja", LangJA},
		{"JA", LangJA},
		{"japanese", LangJA},
		{"jp", LangJA},
		{"nihongo", LangJA},
		{"japones", LangJA},
		{"japonês", LangJA},

		// Chinese
		{"zh", LangZH},
		{"ZH", LangZH},
		{"chinese", LangZH},
		{"mandarin", LangZH},
		{"zh-cn", LangZH},
		{"cn", LangZH},
		{"chines", LangZH},
		{"chinês", LangZH},

		// Fallback
		{"unknown", LangPTBR},
		{"", LangPTBR},
	}

	for _, tt := range tests {
		got := ParseLanguage(tt.input)
		if got != tt.expected {
			t.Errorf("ParseLanguage(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestAllLanguagesHaveTranslations(t *testing.T) {
	for _, langInfo := range AvailableLanguages {
		tr := Get(langInfo.Code)

		if tr.Title == "" {
			t.Errorf("Language %q has empty Title", langInfo.Code)
		}
		if tr.ControlsHeader == "" {
			t.Errorf("Language %q has empty ControlsHeader", langInfo.Code)
		}
		if tr.GameOverTitle == "" {
			t.Errorf("Language %q has empty GameOverTitle", langInfo.Code)
		}
		if tr.PausedTitle == "" {
			t.Errorf("Language %q has empty PausedTitle", langInfo.Code)
		}
		if tr.AIMenuTitle == "" {
			t.Errorf("Language %q has empty AIMenuTitle", langInfo.Code)
		}
		if tr.LangMenuTitle == "" {
			t.Errorf("Language %q has empty LangMenuTitle", langInfo.Code)
		}
		if tr.TerminalTooSmallTitle == "" {
			t.Errorf("Language %q has empty TerminalTooSmallTitle", langInfo.Code)
		}

		// Test Formatters
		clean := tr.FormatAutoPlayCleanup(60)
		if !strings.Contains(clean, "60%") {
			t.Errorf("Language %q cleanup format failed: %q", langInfo.Code, clean)
		}
		look := tr.FormatLookahead(10)
		if !strings.Contains(look, "10") {
			t.Errorf("Language %q lookahead format failed: %q", langInfo.Code, look)
		}
		sim := tr.FormatSimHorizon(50, "85.2%")
		if !strings.Contains(sim, "P50") || !strings.Contains(sim, "85.2%") {
			t.Errorf("Language %q sim horizon format failed: %q", langInfo.Code, sim)
		}
		score := tr.FormatGameOverScore(1234)
		if !strings.Contains(score, "1234") {
			t.Errorf("Language %q score format failed: %q", langInfo.Code, score)
		}
	}
}

func TestDefaultFallback(t *testing.T) {
	tr := Get("non-existent")
	ptTr := Get(LangPTBR)
	if tr.ControlsHeader != ptTr.ControlsHeader {
		t.Errorf("Expected fallback to pt-br, got %q vs %q", tr.ControlsHeader, ptTr.ControlsHeader)
	}
}
