package i18n

import (
	"fmt"
	"strings"
)

// Language represents a supported UI language code.
type Language string

const (
	LangPTBR Language = "pt-br"
	LangEN   Language = "en"
	LangES   Language = "es"
	LangFR   Language = "fr"
	LangIT   Language = "it"
	LangDE   Language = "de"
	LangRU   Language = "ru"
	LangJA   Language = "ja"
	LangZH   Language = "zh"
)

// LanguageInfo represents descriptive metadata about a language.
type LanguageInfo struct {
	Code        Language
	Name        string
	NativeName  string
	Flag        string
}

// AvailableLanguages lists all supported languages in display order.
var AvailableLanguages = []LanguageInfo{
	{Code: LangPTBR, Name: "Portuguese (BR)", NativeName: "Português (Brasil)", Flag: "🇧🇷"},
	{Code: LangEN, Name: "English", NativeName: "English", Flag: "🇺🇸"},
	{Code: LangES, Name: "Spanish", NativeName: "Español", Flag: "🇪🇸"},
	{Code: LangFR, Name: "French", NativeName: "Français", Flag: "🇫🇷"},
	{Code: LangIT, Name: "Italian", NativeName: "Italiano", Flag: "🇮🇹"},
	{Code: LangDE, Name: "German", NativeName: "Deutsch", Flag: "🇩🇪"},
	{Code: LangRU, Name: "Russian", NativeName: "Русский", Flag: "🇷🇺"},
	{Code: LangJA, Name: "Japanese", NativeName: "日本語", Flag: "🇯🇵"},
	{Code: LangZH, Name: "Chinese", NativeName: "简体中文", Flag: "🇨🇳"},
}

// Translations contains all localized strings used across the game.
type Translations struct {
	// Game Title & Badges
	Title           string
	AutoPlayOn      string
	AutoPlayCleanup string // format: "🤖 AUTO-PLAY [🚨 LIMPEZA %d%%+]"
	AutoPlaySpeed   string // format: "🤖 AUTO-PLAY ON [⚡ TETRIS 40%]"
	LookaheadPieces string // format: " [%d peças]"
	HybridBadge     string // format: " [HÍBRIDA]"

	// Left Panel (Stats & Hold)
	HoldHeader string
	Score      string
	HighScore  string
	Level      string
	Lines      string
	Tetris     string
	SimHorizon string // format: "P%d sim: %s"

	// Right Panel (Controls & Next)
	NextHeader     string
	ControlsHeader string
	KeyMove        string
	KeySoftDrop    string
	KeyHardDrop    string
	KeyRotateCW    string
	KeyRotateCCW   string
	KeyHold        string
	KeyAutoPlay    string
	KeySelectAI    string
	KeySelectLang  string
	KeyFourLines   string
	KeySetupTetris string
	KeyPause       string
	KeyRestart     string
	KeyQuit        string

	// Overlays
	PausedTitle     string
	PausedPrompt    string
	GameOverTitle   string
	GameOverScore   string // format: "Pontos: %d"
	GameOverLines   string // format: "Linhas: %d"
	GameOverTetris  string // format: "Tetris: %d"
	GameOverRestart string
	GameOverQuit    string

	// AI Menu
	AIMenuTitle        string
	AIOptionCurrent    string
	AIOptionLookahead  string
	AIOptionHybrid     string
	AIUnavailable      string
	AIMenuHelp         string
	AIModelHorizon     string // format: "Modelo: sobrevivência por %d colocações simuladas."
	AIModelUnavailable string

	// Language Menu
	LangMenuTitle string
	LangMenuHelp  string

	// Terminal Too Small
	TerminalTooSmallTitle string
	TerminalRequirement   string
	TerminalRequired      string // format: "  Exigido: %s"
	TerminalCurrent       string // format: "  Atual:   %s"
	TerminalResizePrompt  string
}

var translations = map[Language]Translations{
	LangPTBR: {
		Title:                 "🎮  T E T R I S   G O  🎮",
		AutoPlayOn:            "🤖 AUTO-PLAY ON",
		AutoPlayCleanup:       "🤖 AUTO-PLAY [🚨 LIMPEZA %d%%+]",
		AutoPlaySpeed:         "🤖 AUTO-PLAY ON [⚡ TETRIS 40%]",
		LookaheadPieces:       " [%d peças]",
		HybridBadge:           " [HÍBRIDA]",
		HoldHeader:            "HOLD [C]",
		Score:                 "SCORE",
		HighScore:             "HIGH SCORE",
		Level:                 "LEVEL",
		Lines:                 "LINES",
		Tetris:                "TETRIS",
		SimHorizon:            "P%d sim: %s",
		NextHeader:            "NEXT",
		ControlsHeader:        "CONTROLES",
		KeyMove:               "Mover",
		KeySoftDrop:           "Soft Drop",
		KeyHardDrop:           "Hard Drop",
		KeyRotateCW:           "Girar Horário",
		KeyRotateCCW:          "Girar Anti-h",
		KeyHold:               "Guardar Peça",
		KeyAutoPlay:           "Auto-Play (IA)",
		KeySelectAI:           "Escolher IA",
		KeySelectLang:         "Idioma",
		KeyFourLines:          "4 Linhas (I)",
		KeySetupTetris:        "Setup Tetris",
		KeyPause:              "Pausar",
		KeyRestart:            "Reiniciar",
		KeyQuit:               "Sair",
		PausedTitle:           "   PAUSADO   ",
		PausedPrompt:          "Pressione P\npara continuar",
		GameOverTitle:         " GAME OVER ",
		GameOverScore:         "Pontos: %d",
		GameOverLines:         "Linhas: %d",
		GameOverTetris:        "Tetris: %d",
		GameOverRestart:       "[R] Reiniciar",
		GameOverQuit:          "[Q] Sair",
		AIMenuTitle:           "ESCOLHER IA",
		AIOptionCurrent:       "1. IA atual (v2)",
		AIOptionLookahead:     "2. Lookahead 10 peças (beam 4)",
		AIOptionHybrid:        "3. IA híbrida (experimental)",
		AIUnavailable:         " — indisponível",
		AIMenuHelp:            "↑/↓ ou 1/2/3: escolher\nEnter: confirmar | Esc/M: voltar\n\nAlterar IA reinicia a partida.\nB/Tab: ativar Auto-Play no jogo.",
		AIModelHorizon:        "Modelo: sobrevivência por %d colocações simuladas.",
		AIModelUnavailable:    "MODELO INDISPONÍVEL",
		LangMenuTitle:         "ESCOLHER IDIOMA",
		LangMenuHelp:          "↑/↓ ou 1..9: escolher\nEnter: confirmar | Esc/L: voltar",
		TerminalTooSmallTitle: "⚠️  TERMINAL MUITO PEQUENO  ⚠️",
		TerminalRequirement:   "Para uma exibição adequada do Tetris:",
		TerminalRequired:      "  Exigido: %s",
		TerminalCurrent:       "  Atual:   %s",
		TerminalResizePrompt:  "Por favor, aumente o tamanho da sua janela.",
	},
	LangEN: {
		Title:                 "🎮  T E T R I S   G O  🎮",
		AutoPlayOn:            "🤖 AUTO-PLAY ON",
		AutoPlayCleanup:       "🤖 AUTO-PLAY [🚨 CLEANUP %d%%+]",
		AutoPlaySpeed:         "🤖 AUTO-PLAY ON [⚡ TETRIS 40%]",
		LookaheadPieces:       " [%d pieces]",
		HybridBadge:           " [HYBRID]",
		HoldHeader:            "HOLD [C]",
		Score:                 "SCORE",
		HighScore:             "HIGH SCORE",
		Level:                 "LEVEL",
		Lines:                 "LINES",
		Tetris:                "TETRIS",
		SimHorizon:            "P%d sim: %s",
		NextHeader:            "NEXT",
		ControlsHeader:        "CONTROLS",
		KeyMove:               "Move",
		KeySoftDrop:           "Soft Drop",
		KeyHardDrop:           "Hard Drop",
		KeyRotateCW:           "Rotate CW",
		KeyRotateCCW:          "Rotate CCW",
		KeyHold:               "Hold Piece",
		KeyAutoPlay:           "Auto-Play (AI)",
		KeySelectAI:           "Select AI",
		KeySelectLang:         "Language",
		KeyFourLines:          "4 Lines (I)",
		KeySetupTetris:        "Setup Tetris",
		KeyPause:              "Pause",
		KeyRestart:            "Restart",
		KeyQuit:               "Quit",
		PausedTitle:           "   PAUSED   ",
		PausedPrompt:          "Press P\nto continue",
		GameOverTitle:         " GAME OVER ",
		GameOverScore:         "Score: %d",
		GameOverLines:         "Lines: %d",
		GameOverTetris:        "Tetris: %d",
		GameOverRestart:       "[R] Restart",
		GameOverQuit:          "[Q] Quit",
		AIMenuTitle:           "SELECT AI",
		AIOptionCurrent:       "1. Current AI (v2)",
		AIOptionLookahead:     "2. Lookahead 10 pieces (beam 4)",
		AIOptionHybrid:        "3. Hybrid AI (experimental)",
		AIUnavailable:         " — unavailable",
		AIMenuHelp:            "↑/↓ or 1/2/3: choose\nEnter: confirm | Esc/M: back\n\nChanging AI restarts the game.\nB/Tab: toggle Auto-Play.",
		AIModelHorizon:        "Model: survival over %d simulated placements.",
		AIModelUnavailable:    "MODEL UNAVAILABLE",
		LangMenuTitle:         "SELECT LANGUAGE",
		LangMenuHelp:          "↑/↓ or 1..9: choose\nEnter: confirm | Esc/L: back",
		TerminalTooSmallTitle: "⚠️  TERMINAL TOO SMALL  ⚠️",
		TerminalRequirement:   "For a proper display of Tetris:",
		TerminalRequired:      "  Required: %s",
		TerminalCurrent:       "  Current:  %s",
		TerminalResizePrompt:  "Please increase your window size.",
	},
	LangES: {
		Title:                 "🎮  T E T R I S   G O  🎮",
		AutoPlayOn:            "🤖 AUTO-PLAY ON",
		AutoPlayCleanup:       "🤖 AUTO-PLAY [🚨 LIMPIEZA %d%%+] ",
		AutoPlaySpeed:         "🤖 AUTO-PLAY ON [⚡ TETRIS 40%]",
		LookaheadPieces:       " [%d piezas]",
		HybridBadge:           " [HÍBRIDA]",
		HoldHeader:            "HOLD [C]",
		Score:                 "PUNTOS",
		HighScore:             "RÉCORD",
		Level:                 "NIVEL",
		Lines:                 "LÍNEAS",
		Tetris:                "TETRIS",
		SimHorizon:            "P%d sim: %s",
		NextHeader:            "SIGUIENTE",
		ControlsHeader:        "CONTROLES",
		KeyMove:               "Mover",
		KeySoftDrop:           "Caída Lenta",
		KeyHardDrop:           "Caída Rápida",
		KeyRotateCW:           "Girar Horario",
		KeyRotateCCW:          "Girar Anti-h",
		KeyHold:               "Guardar Pieza",
		KeyAutoPlay:           "Auto-Play (IA)",
		KeySelectAI:           "Elegir IA",
		KeySelectLang:         "Idioma",
		KeyFourLines:          "4 Líneas (I)",
		KeySetupTetris:        "Setup Tetris",
		KeyPause:              "Pausar",
		KeyRestart:            "Reiniciar",
		KeyQuit:               "Salir",
		PausedTitle:           "   PAUSADO   ",
		PausedPrompt:          "Presiona P\npara continuar",
		GameOverTitle:         " GAME OVER ",
		GameOverScore:         "Puntos: %d",
		GameOverLines:         "Líneas: %d",
		GameOverTetris:        "Tetris: %d",
		GameOverRestart:       "[R] Reiniciar",
		GameOverQuit:          "[Q] Salir",
		AIMenuTitle:           "ELEGIR IA",
		AIOptionCurrent:       "1. IA actual (v2)",
		AIOptionLookahead:     "2. Lookahead 10 piezas (beam 4)",
		AIOptionHybrid:        "3. IA híbrida (experimental)",
		AIUnavailable:         " — no disponible",
		AIMenuHelp:            "↑/↓ o 1/2/3: elegir\nEnter: confirmar | Esc/M: volver\n\nCambiar IA reinicia la partida.\nB/Tab: activar Auto-Play en el juego.",
		AIModelHorizon:        "Modelo: supervivencia por %d colocaciones simuladas.",
		AIModelUnavailable:    "MODELO NO DISPONIBLE",
		LangMenuTitle:         "ELEGIR IDIOMA",
		LangMenuHelp:          "↑/↓ o 1..9: elegir\nEnter: confirmar | Esc/L: volver",
		TerminalTooSmallTitle: "⚠️  TERMINAL MUY PEQUEÑO  ⚠️",
		TerminalRequirement:   "Para una correcta visualização de Tetris:",
		TerminalRequired:      "  Requerido: %s",
		TerminalCurrent:       "  Actual:    %s",
		TerminalResizePrompt:  "Por favor, aumenta el tamaño de tu ventana.",
	},
	LangFR: {
		Title:                 "🎮  T E T R I S   G O  🎮",
		AutoPlayOn:            "🤖 AUTO-PLAY ON",
		AutoPlayCleanup:       "🤖 AUTO-PLAY [🚨 NETTOYAGE %d%%+]",
		AutoPlaySpeed:         "🤖 AUTO-PLAY ON [⚡ TETRIS 40%]",
		LookaheadPieces:       " [%d pièces]",
		HybridBadge:           " [HYBRIDE]",
		HoldHeader:            "HOLD [C]",
		Score:                 "SCORE",
		HighScore:             "MEILLEUR",
		Level:                 "NIVEAU",
		Lines:                 "LIGNES",
		Tetris:                "TETRIS",
		SimHorizon:            "P%d sim: %s",
		NextHeader:            "SUIVANT",
		ControlsHeader:        "COMMANDES",
		KeyMove:               "Déplacer",
		KeySoftDrop:           "Chute Lente",
		KeyHardDrop:           "Chute Rapide",
		KeyRotateCW:           "Rotation Hor.",
		KeyRotateCCW:          "Rotation Anti-h",
		KeyHold:               "Garder Pièce",
		KeyAutoPlay:           "Auto-Play (IA)",
		KeySelectAI:           "Choisir IA",
		KeySelectLang:         "Langue",
		KeyFourLines:          "4 Lignes (I)",
		KeySetupTetris:        "Setup Tetris",
		KeyPause:              "Pause",
		KeyRestart:            "Recommencer",
		KeyQuit:               "Quitter",
		PausedTitle:           "   EN PAUSE   ",
		PausedPrompt:          "Appuyez sur P\npour continuer",
		GameOverTitle:         " GAME OVER ",
		GameOverScore:         "Score: %d",
		GameOverLines:         "Lignes: %d",
		GameOverTetris:        "Tetris: %d",
		GameOverRestart:       "[R] Recommencer",
		GameOverQuit:          "[Q] Quitter",
		AIMenuTitle:           "CHOISIR IA",
		AIOptionCurrent:       "1. IA actuelle (v2)",
		AIOptionLookahead:     "2. Lookahead 10 pièces (beam 4)",
		AIOptionHybrid:        "3. IA hybride (expérimentale)",
		AIUnavailable:         " — indisponible",
		AIMenuHelp:            "↑/↓ ou 1/2/3: choisir\nEnter: confirmer | Esc/M: retour\n\nChanger l'IA redémarre la partie.\nB/Tab: basculer l'Auto-Play.",
		AIModelHorizon:        "Modèle: survie sur %d placements simulés.",
		AIModelUnavailable:    "MODÈLE INDISPONIBLE",
		LangMenuTitle:         "CHOISIR LA LANGUE",
		LangMenuHelp:          "↑/↓ ou 1..9: choisir\nEnter: confirmer | Esc/L: retour",
		TerminalTooSmallTitle: "⚠️  TERMINAL TROP PETIT  ⚠️",
		TerminalRequirement:   "Pour un affichage correct de Tetris :",
		TerminalRequired:      "  Requis:  %s",
		TerminalCurrent:       "  Actuel:  %s",
		TerminalResizePrompt:  "Veuillez agrandir la taille de votre fenêtre.",
	},
	LangIT: {
		Title:                 "🎮  T E T R I S   G O  🎮",
		AutoPlayOn:            "🤖 AUTO-PLAY ON",
		AutoPlayCleanup:       "🤖 AUTO-PLAY [🚨 PULIZIA %d%%+]",
		AutoPlaySpeed:         "🤖 AUTO-PLAY ON [⚡ TETRIS 40%]",
		LookaheadPieces:       " [%d pezzi]",
		HybridBadge:           " [IBRIDO]",
		HoldHeader:            "HOLD [C]",
		Score:                 "PUNTI",
		HighScore:             "RECORD",
		Level:                 "LIVELLO",
		Lines:                 "LINEE",
		Tetris:                "TETRIS",
		SimHorizon:            "P%d sim: %s",
		NextHeader:            "PROSSIMO",
		ControlsHeader:        "CONTROLLI",
		KeyMove:               "Muovi",
		KeySoftDrop:           "Caduta Lenta",
		KeyHardDrop:           "Caduta Rapida",
		KeyRotateCW:           "Ruota Orario",
		KeyRotateCCW:          "Ruota Anti-o",
		KeyHold:               "Trattieni Pezzo",
		KeyAutoPlay:           "Auto-Play (IA)",
		KeySelectAI:           "Scegli IA",
		KeySelectLang:         "Lingua",
		KeyFourLines:          "4 Linee (I)",
		KeySetupTetris:        "Setup Tetris",
		KeyPause:              "Pausa",
		KeyRestart:            "Riavvia",
		KeyQuit:               "Esci",
		PausedTitle:           "   IN PAUSA   ",
		PausedPrompt:          "Premi P\nper continuare",
		GameOverTitle:         " GAME OVER ",
		GameOverScore:         "Punti: %d",
		GameOverLines:         "Linee: %d",
		GameOverTetris:        "Tetris: %d",
		GameOverRestart:       "[R] Riavvia",
		GameOverQuit:          "[Q] Esci",
		AIMenuTitle:           "SCEGLI IA",
		AIOptionCurrent:       "1. IA attuale (v2)",
		AIOptionLookahead:     "2. Lookahead 10 pezzi (beam 4)",
		AIOptionHybrid:        "3. IA ibrida (sperimentale)",
		AIUnavailable:         " — non disponibile",
		AIMenuHelp:            "↑/↓ o 1/2/3: scegliere\nEnter: conferma | Esc/M: indietro\n\nCambiare IA riavvia la partita.\nB/Tab: attiva Auto-Play nel gioco.",
		AIModelHorizon:        "Modello: sopravvivenza per %d posizionamenti simulati.",
		AIModelUnavailable:    "MODELLO NON DISPONIBILE",
		LangMenuTitle:         "SCEGLI LINGUA",
		LangMenuHelp:          "↑/↓ o 1..9: scegliere\nEnter: conferma | Esc/L: indietro",
		TerminalTooSmallTitle: "⚠️  TERMINAL TROPPO PICCOLO  ⚠️",
		TerminalRequirement:   "Per una corretta visualizzazione di Tetris:",
		TerminalRequired:      "  Richiesto: %s",
		TerminalCurrent:       "  Attuale:   %s",
		TerminalResizePrompt:  "Si prega di aumentare le dimensioni della finestra.",
	},
	LangDE: {
		Title:                 "🎮  T E T R I S   G O  🎮",
		AutoPlayOn:            "🤖 AUTO-PLAY EIN",
		AutoPlayCleanup:       "🤖 AUTO-PLAY [🚨 BEREINIGUNG %d%%+]",
		AutoPlaySpeed:         "🤖 AUTO-PLAY EIN [⚡ TETRIS 40%]",
		LookaheadPieces:       " [%d Teile]",
		HybridBadge:           " [HYBRID]",
		HoldHeader:            "HOLD [C]",
		Score:                 "PUNKTE",
		HighScore:             "REKORD",
		Level:                 "LEVEL",
		Lines:                 "REIHEN",
		Tetris:                "TETRIS",
		SimHorizon:            "P%d Sim: %s",
		NextHeader:            "NÄCHSTE",
		ControlsHeader:        "STEUERUNG",
		KeyMove:               "Bewegen",
		KeySoftDrop:           "Soft Drop",
		KeyHardDrop:           "Hard Drop",
		KeyRotateCW:           "Drehen UZS",
		KeyRotateCCW:          "Drehen GZS",
		KeyHold:               "Halten",
		KeyAutoPlay:           "Auto-Play (KI)",
		KeySelectAI:           "KI Wählen",
		KeySelectLang:         "Sprache",
		KeyFourLines:          "4 Reihen (I)",
		KeySetupTetris:        "Setup Tetris",
		KeyPause:              "Pause",
		KeyRestart:            "Neustart",
		KeyQuit:               "Beenden",
		PausedTitle:           "   PAUSIERT   ",
		PausedPrompt:          "Drücke P\nzum Fortsetzen",
		GameOverTitle:         " GAME OVER ",
		GameOverScore:         "Punkte: %d",
		GameOverLines:         "Reihen: %d",
		GameOverTetris:        "Tetris: %d",
		GameOverRestart:       "[R] Neustart",
		GameOverQuit:          "[Q] Beenden",
		AIMenuTitle:           "KI WÄHLEN",
		AIOptionCurrent:       "1. Aktuelle KI (v2)",
		AIOptionLookahead:     "2. Lookahead 10 Teile (Beam 4)",
		AIOptionHybrid:        "3. Hybride KI (experimentell)",
		AIUnavailable:         " — nicht verfügbar",
		AIMenuHelp:            "↑/↓ oder 1/2/3: wählen\nEnter: bestätigen | Esc/M: zurück\n\nKI-Wechsel startet das Spiel neu.\nB/Tab: Auto-Play umschalten.",
		AIModelHorizon:        "Modell: Überleben über %d simulierte Platzierungen.",
		AIModelUnavailable:    "MODELL NICHT VERFÜGBAR",
		LangMenuTitle:         "SPRACHE WÄHLEN",
		LangMenuHelp:          "↑/↓ oder 1..9: wählen\nEnter: bestätigen | Esc/L: zurück",
		TerminalTooSmallTitle: "⚠️  TERMINAL ZU KLEIN  ⚠️",
		TerminalRequirement:   "Für eine korrekte Darstellung von Tetris:",
		TerminalRequired:      "  Erforderlich: %s",
		TerminalCurrent:       "  Aktuell:      %s",
		TerminalResizePrompt:  "Bitte vergrößere dein Terminalfenster.",
	},
	LangRU: {
		Title:                 "🎮  T E T R I S   G O  🎮",
		AutoPlayOn:            "🤖 АВТО-ИГРА ВКЛ",
		AutoPlayCleanup:       "🤖 АВТО-ИГРА [🚨 ОЧИСТКА %d%%+]",
		AutoPlaySpeed:         "🤖 АВТО-ИГРА ВКЛ [⚡ ТЕТРИС 40%] ",
		LookaheadPieces:       " [%d фигур]",
		HybridBadge:           " [ГИБРИД]",
		HoldHeader:            "HOLD [C]",
		Score:                 "ОЧКИ",
		HighScore:             "РЕКОРД",
		Level:                 "УРОВЕНЬ",
		Lines:                 "ЛИНИИ",
		Tetris:                "ТЕТРИС",
		SimHorizon:            "P%d сим: %s",
		NextHeader:            "ДАЛЕЕ",
		ControlsHeader:        "УПРАВЛЕНИЕ",
		KeyMove:               "Движение",
		KeySoftDrop:           "Soft Drop",
		KeyHardDrop:           "Хард-дроп",
		KeyRotateCW:           "По часовой",
		KeyRotateCCW:          "Против час.",
		KeyHold:               "Удержание",
		KeyAutoPlay:           "Авто-игра (ИИ)",
		KeySelectAI:           "Выбор ИИ",
		KeySelectLang:         "Язык",
		KeyFourLines:          "4 Линии (I)",
		KeySetupTetris:        "Сетап Тетрис",
		KeyPause:              "Пауза",
		KeyRestart:            "Заново",
		KeyQuit:               "Выход",
		PausedTitle:           "   ПАУЗА   ",
		PausedPrompt:          "Нажмите P\nдля продолжения",
		GameOverTitle:         " GAME OVER ",
		GameOverScore:         "Очки: %d",
		GameOverLines:         "Линии: %d",
		GameOverTetris:        "Тетрис: %d",
		GameOverRestart:       "[R] Заново",
		GameOverQuit:          "[Q] Выход",
		AIMenuTitle:           "ВЫБОР ИИ",
		AIOptionCurrent:       "1. Текущий ИИ (v2)",
		AIOptionLookahead:     "2. Lookahead 10 фигур (beam 4)",
		AIOptionHybrid:        "3. Гибридный ИИ (эксперимент)",
		AIUnavailable:         " — недоступен",
		AIMenuHelp:            "↑/↓ или 1/2/3: выбор\nEnter: подтвердить | Esc/M: назад\n\nСмена ИИ перезапускает игру.\nB/Tab: переключить авто-игру.",
		AIModelHorizon:        "Модель: выживание на протяжении %d симуляций.",
		AIModelUnavailable:    "МОДЕЛЬ НЕДОСТУПНА",
		LangMenuTitle:         "ВЫБОР ЯЗЫКА",
		LangMenuHelp:          "↑/↓ или 1..9: выбор\nEnter: подтвердить | Esc/L: назад",
		TerminalTooSmallTitle: "⚠️  ТЕРМИНАЛ СЛИШКОМ МАЛ  ⚠️",
		TerminalRequirement:   "Для корректного отображения Тетриса:",
		TerminalRequired:      "  Требуется: %s",
		TerminalCurrent:       "  Текущий:   %s",
		TerminalResizePrompt:  "Пожалуйста, увеличьте размер окна.",
	},
	LangJA: {
		Title:                 "🎮  T E T R I S   G O  🎮",
		AutoPlayOn:            "🤖 オートプレイ ON",
		AutoPlayCleanup:       "🤖 オートプレイ [🚨 緊急消去 %d%%+]",
		AutoPlaySpeed:         "🤖 オートプレイ ON [⚡ テトリス 40%]",
		LookaheadPieces:       " [%dミノ先]",
		HybridBadge:           " [ハイブリッド]",
		HoldHeader:            "HOLD [C]",
		Score:                 "スコア",
		HighScore:             "ハイスコア",
		Level:                 "レベル",
		Lines:                 "ライン",
		Tetris:                "テトリス",
		SimHorizon:            "P%d 予測: %s",
		NextHeader:            "NEXT",
		ControlsHeader:        "操作方法",
		KeyMove:               "移動",
		KeySoftDrop:           "ソフト落下",
		KeyHardDrop:           "ハードドロップ",
		KeyRotateCW:           "右回転",
		KeyRotateCCW:          "左回転",
		KeyHold:               "ホールド",
		KeyAutoPlay:           "AI自動プレイ",
		KeySelectAI:           "AI選択",
		KeySelectLang:         "言語切替",
		KeyFourLines:          "I型4連続 (I)",
		KeySetupTetris:        "テトリス準備",
		KeyPause:              "一時停止",
		KeyRestart:            "やり直し",
		KeyQuit:               "終了",
		PausedTitle:           "   ポーズ中   ",
		PausedPrompt:          "Pキーを押して\n再開",
		GameOverTitle:         " GAME OVER ",
		GameOverScore:         "スコア: %d",
		GameOverLines:         "ライン: %d",
		GameOverTetris:        "テトリス: %d",
		GameOverRestart:       "[R] もう一度",
		GameOverQuit:          "[Q] 終了",
		AIMenuTitle:           "AI選択",
		AIOptionCurrent:       "1. 現在のAI (v2)",
		AIOptionLookahead:     "2. 先読み10ミノ (beam 4)",
		AIOptionHybrid:        "3. ハイブリッドAI (実験的)",
		AIUnavailable:         " — 利用不可",
		AIMenuHelp:            "↑/↓ または 1/2/3: 選択\nEnter: 確定 | Esc/M: 戻る\n\nAI変更時はゲームが再スタートします。\nB/Tab: オートプレイ切替",
		AIModelHorizon:        "モデル: %d回のシミュレーションで生存検証済み",
		AIModelUnavailable:    "モデル利用不可",
		LangMenuTitle:         "言語選択",
		LangMenuHelp:          "↑/↓ または 1..9: 選択\nEnter: 確定 | Esc/L: 戻る",
		TerminalTooSmallTitle: "⚠️  画面サイズが小さすぎます  ⚠️",
		TerminalRequirement:   "テトリスを正常に表示するには:",
		TerminalRequired:      "  必要サイズ: %s",
		TerminalCurrent:       "  現在サイズ: %s",
		TerminalResizePrompt:  "ターミナルのウィンドウを拡大してください。",
	},
	LangZH: {
		Title:                 "🎮  T E T R I S   G O  🎮",
		AutoPlayOn:            "🤖 自动托管 ON",
		AutoPlayCleanup:       "🤖 自动托管 [🚨 紧急清理 %d%%+]",
		AutoPlaySpeed:         "🤖 自动托管 ON [⚡ 消除4行 40%]",
		LookaheadPieces:       " [%d块预判]",
		HybridBadge:           " [混合AI]",
		HoldHeader:            "HOLD [C]",
		Score:                 "得分",
		HighScore:             "最高分",
		Level:                 "等级",
		Lines:                 "消行",
		Tetris:                "四行消",
		SimHorizon:            "P%d 预测: %s",
		NextHeader:            "NEXT",
		ControlsHeader:        "操作说明",
		KeyMove:               "左右移动",
		KeySoftDrop:           "加速下落",
		KeyHardDrop:           "直接下落",
		KeyRotateCW:           "顺时针旋转",
		KeyRotateCCW:          "逆时针旋转",
		KeyHold:               "暂存方块",
		KeyAutoPlay:           "AI自动托管",
		KeySelectAI:           "选择AI算法",
		KeySelectLang:         "语言设置",
		KeyFourLines:          "连续4个条 (I)",
		KeySetupTetris:        "快速摆盘",
		KeyPause:              "暂停",
		KeyRestart:            "重新开始",
		KeyQuit:               "退出游戏",
		PausedTitle:           "   已暂停   ",
		PausedPrompt:          "按 P 键\n继续游戏",
		GameOverTitle:         " GAME OVER ",
		GameOverScore:         "得分: %d",
		GameOverLines:         "消除行数: %d",
		GameOverTetris:        "四行消除: %d",
		GameOverRestart:       "[R] 重新开始",
		GameOverQuit:          "[Q] 退出",
		AIMenuTitle:           "选择AI算法",
		AIOptionCurrent:       "1. 当前AI (v2)",
		AIOptionLookahead:     "2. 预判10块 (beam 4)",
		AIOptionHybrid:        "3. 混合AI模型 (实验性)",
		AIUnavailable:         " — 不可用",
		AIMenuHelp:            "↑/↓ 或 1/2/3: 选择\nEnter: 确认 | Esc/M: 返回\n\n切换AI将重新开始游戏。\nB/Tab: 切换自动托管",
		AIModelHorizon:        "模型: 经过 %d 次模拟落点生存训练",
		AIModelUnavailable:    "模型不可用",
		LangMenuTitle:         "选择语言",
		LangMenuHelp:          "↑/↓ 或 1..9: 选择\nEnter: 确认 | Esc/L: 返回",
		TerminalTooSmallTitle: "⚠️  终端窗口过小  ⚠️",
		TerminalRequirement:   "为了正常显示俄罗斯方块:",
		TerminalRequired:      "  需要尺寸: %s",
		TerminalCurrent:       "  当前尺寸: %s",
		TerminalResizePrompt:  "请调大终端窗口尺寸。",
	},
}

// Get returns the Translations for the requested Language, defaulting to pt-br.
func Get(lang Language) Translations {
	if t, ok := translations[lang]; ok {
		return t
	}
	return translations[LangPTBR]
}

// ParseLanguage normalizes user input string into a supported Language.
func ParseLanguage(s string) Language {
	norm := strings.ToLower(strings.TrimSpace(s))
	norm = strings.ReplaceAll(norm, "_", "-")
	switch norm {
	case "en", "english", "ing", "ingles", "inglês":
		return LangEN
	case "es", "spanish", "espanol", "español", "esp":
		return LangES
	case "fr", "french", "francais", "français", "fra":
		return LangFR
	case "it", "ita", "italian", "italiano":
		return LangIT
	case "de", "german", "deutsch", "ger", "alemao", "alemão":
		return LangDE
	case "ru", "russian", "rus", "russo":
		return LangRU
	case "ja", "japanese", "jp", "nihongo", "japones", "japonês":
		return LangJA
	case "zh", "chinese", "mandarin", "zh-cn", "cn", "chines", "chinês":
		return LangZH
	case "pt", "pt-br", "ptbr", "portugues", "português", "portuguese":
		return LangPTBR
	default:
		return LangPTBR
	}
}

// FormatAutoPlayCleanup renders the cleanup threshold badge for the given language.
func (t Translations) FormatAutoPlayCleanup(percent int) string {
	return fmt.Sprintf(t.AutoPlayCleanup, percent)
}

// FormatLookahead renders the lookahead piece count for the badge.
func (t Translations) FormatLookahead(depth int) string {
	return fmt.Sprintf(t.LookaheadPieces, depth)
}

// FormatSimHorizon renders the model simulation horizon and probability.
func (t Translations) FormatSimHorizon(horizon int, prob string) string {
	return fmt.Sprintf(t.SimHorizon, horizon, prob)
}

// FormatGameOverScore renders the game over score text.
func (t Translations) FormatGameOverScore(score int) string {
	return fmt.Sprintf(t.GameOverScore, score)
}

// FormatGameOverLines renders the game over lines text.
func (t Translations) FormatGameOverLines(lines int) string {
	return fmt.Sprintf(t.GameOverLines, lines)
}

// FormatGameOverTetris renders the game over tetris text.
func (t Translations) FormatGameOverTetris(tetrises int) string {
	return fmt.Sprintf(t.GameOverTetris, tetrises)
}

// FormatAIModelHorizon renders the model horizon description in AI menu.
func (t Translations) FormatAIModelHorizon(horizon int) string {
	return fmt.Sprintf(t.AIModelHorizon, horizon)
}
