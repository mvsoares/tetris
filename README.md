# 🎮 Tetris CLI in Go

[![Release](https://img.shields.io/badge/release-v1.0.0-blue.svg)](https://github.com/mvsoares/tetris/releases)
[![Go Version](https://img.shields.io/badge/go-1.26+-00ADD8.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![i18n](https://img.shields.io/badge/languages-9%20Languages%20(PT--BR%20%7C%20EN%20%7C%20ES%20%7C%20FR%20%7C%20IT%20%7C%20DE%20%7C%20RU%20%7C%20JA%20%7C%20ZH)-orange.svg)](#-languages--multilingual-support)

*Translations: [🇺🇸 English](README.md) | [🇧🇷 Português](README.pt-br.md)*

A modern, fluid, and high-performance terminal implementation of the classic **Tetris**, built in **Go** using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) architecture (The Elm Architecture) and styled with [Lipgloss](https://github.com/charmbracelet/lipgloss).

Features **native support for 9 languages** (Portuguese, English, Spanish, French, Italian, German, Russian, Japanese, and Simplified Chinese), **Auto-Play powered by an Adaptive Heuristic AI & BitBoard Engine**, **a concurrent headless simulation engine**, an **asynchronous logger with a 100 MB buffer** for JSONL dataset generation, and CLI tools for deep game and loss analysis.

See the [changelog](CHANGELOG.md) and [benchmark results](benchmarks/README.md).

---

## 🌐 Languages / Multilingual Support

Tetris includes full, native translation for **9 languages**:
- 🇧🇷 **Português do Brasil** (`pt-br`, `pt`)
- 🇺🇸 **English** (`en`, `english`) — *international*
- 🇪🇸 **Español** (`es`, `spanish`)
- 🇫🇷 **Français** (`fr`, `french`)
- 🇮🇹 **Italiano** (`it`, `ita`, `italian`)
- 🇩🇪 **Deutsch** (`de`, `german`, `deutsch`)
- 🇷🇺 **Русский** (`ru`, `russian`)
- 🇯🇵 **日本語** (`ja`, `japanese`)
- 🇨🇳 **简体中文** (`zh`, `chinese`)

### How to switch languages:
1. **During gameplay**: Press the **`L`** key at any moment to open the interactive language modal (choose using numeric keys **`1` to `9`**, or arrow keys).
2. **From the command line**: Launch the game with the `--lang` or `--idiom` flag:
   ```bash
   tetris --lang=en        # English
   tetris --lang=ja        # 日本語 (Japanese)
   tetris --lang=ru        # Русский (Russian)
   tetris --lang=de        # Deutsch (German)
   tetris --lang=zh        # 简体中文 (Chinese)
   tetris --lang=es        # Español (Spanish)
   tetris --lang=fr        # Français (French)
   tetris --lang=it        # Italiano (Italian)
   tetris --lang=pt-br     # Português do Brasil
   ```

---

## 🌟 Key Features

- 🌐 **Global Multilingual (i18n)**: 9 fully translated languages with runtime hot-swapping (`L`, keys 1..9) and CLI startup flags (`--lang`, `--idiom`).
- 📦 **Official Packaging (v1.0.0)**: Production-ready packages for **Debian/Ubuntu (`.deb`)**, **Fedora/RHEL (`.rpm`)**, **macOS (`.tar.gz` / Homebrew)**, and **Windows (`.zip`)**.
- 🎨 **Rich Terminal Interface**: TrueColor flicker-free rendering in terminal alternate buffer (`tea.WithAltScreen`), 1:2 block aspect ratio (`██`), and responsive window resize handling (`SIGWINCH`).
- 🕹️ **Official Modern Tetris Mechanics**:
  - Fair **7-Bag** randomizer (prevents prolonged piece droughts).
  - Rotation with **Wall Kicks** (SRS-compliant kicks against boundaries and blocks).
  - Translucent **Ghost Piece** projection (`░░`).
  - **Hold Queue** with a strict 1-swap per lock limit.
  - On-screen **Tetris Counter** (side panel and Game Over summary).
  - **Progressive Gravity**: Fall speed increases dynamically every 10 lines cleared.
  - **Lock Delay**: Generous placement tolerance with up to 15 move/rotation resets for human players.
- 🤖 **Intelligent Auto-Play (9-0 Stacking, BitBoard Engine & Lookahead AI)**:
  - **BitBoard Engine**: Stack-allocated 40-byte bitwise board representation (`[20]uint16`) accelerating calculations up to 6.8x.
  - **9-0 Strategy**: Builds the stack across columns 0–8, reserving column 9 for Tetris line clears with the `I` piece.
  - **Gravity-Aware Legal Pathfinding**: Plans real physical move sequences with rotation and recalculates dynamically if blocked.
  - **Adaptive Lookahead**: Heuristic v2 evaluates incoming queue pieces and deepens search during danger states.
  - **Experimental 10-Piece Beam Search**: Multi-ply beam search with hold evaluation and duplicate state pruning via the `M` menu.
  - **Local Risk Model**: Trained logistic model evaluated against simulated rollouts; selectable as an experimental policy.
- ⚡ **Concurrent Headless Simulator (`./train`)**:
  - Worker pool supporting up to **50 concurrent parallel games** utilizing all CPU cores.
  - Fully reproducible seeds and dual simulation modes (`placement` and `gameplay`).
- 💾 **Asynchronous Logger with 100 MB Buffer**:
  - Non-blocking logging pipeline backed by `bufio.Writer` and a 131,072-item buffered channel.
  - Emits JSON Lines (`.jsonl`) with 20×10 grid snapshots, surface contour, hole count, score, and action telemetry.
- 📊 **Analytical CLI Suite**:
  - `./analyze`: Aggregate performance metrics, Tetris clear rate, scoring, and line distributions.
  - `./analyze_losses`: Root-cause forensic analysis of Game Over triggers (top-out piece, column heights, holes, emergency state).

---

## 🕹️ Game Controls

```text
┌──────────────────────┬────────────────────────────────────────────────────────┐
│ Key                  │ Action / Function                                      │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ ← / A / H            │ Move piece Left                                        │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ → / D / L (in-game)  │ Move piece Right                                       │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ ↓ / S / J            │ Soft Drop (faster drop with score bonus)               │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ Space                │ Hard Drop (instant drop and lock)                      │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ ↑ / W / X            │ Rotate Clockwise (CW)                                  │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ Z                    │ Rotate Counter-Clockwise (CCW)                         │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ C / H                │ Hold / Swap piece in Hold Queue                        │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ B / Tab              │ Toggle Auto-Play (AI plays automatically)              │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ M                    │ Open AI Policy selection menu                          │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ L                    │ Open Language selection menu (9 languages)             │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ 4 / I                │ Force 4 consecutive Line (I) pieces (Debug/Practice)   │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ T                    │ Instant 4-Line Tetris Setup (Debug/Practice)           │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ P                    │ Pause / Resume game                                    │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ R                    │ Restart game                                           │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ Q / Esc / Ctrl+C     │ Quit game                                              │
└──────────────────────┴────────────────────────────────────────────────────────┘
```

---

## 📦 Official Package Installation (v1.0.0)

### Option 1: Universal One-Line Installer (Linux & macOS)
Run this command in your terminal to automatically detect your OS, download the appropriate package, and install:
```bash
curl -fsSL https://raw.githubusercontent.com/mvsoares/tetris/main/scripts/install.sh | bash
```

### Option 2: Debian / Ubuntu (`.deb`)
Download the `.deb` release package for your architecture (`amd64` or `arm64`) and install:
```bash
# For x86_64 / amd64:
sudo dpkg -i tetris_1.0.0_amd64.deb

# For ARM64:
sudo dpkg -i tetris_1.0.0_arm64.deb
```

### Option 3: Fedora / Red Hat / CentOS (`.rpm`)
```bash
# For x86_64:
sudo rpm -Uvh tetris-1.0.0-1.x86_64.rpm
# or:
sudo dnf install tetris-1.0.0-1.x86_64.rpm

# For aarch64 (ARM64):
sudo rpm -Uvh tetris-1.0.0-1.aarch64.rpm
```

### Option 4: macOS (Apple Silicon or Intel)
Download the `.tar.gz` archive for your CPU:
```bash
# Apple Silicon (M1/M2/M3/M4):
tar -xzf tetris_1.0.0_darwin_arm64.tar.gz
sudo mv tetris /usr/local/bin/

# Intel Mac:
tar -xzf tetris_1.0.0_darwin_amd64.tar.gz
sudo mv tetris /usr/local/bin/
```
Or install via [Homebrew](Formula/tetris.rb):
```bash
brew install mvsoares/tetris/tetris
```

### Option 5: Windows
Download `tetris_1.0.0_windows_amd64.zip`, extract `tetris.exe`, and run it in Windows Terminal, PowerShell, or Command Prompt.

---

## 📥 Building from Source

### Prerequisites

- **Go 1.26** or higher installed ([go.dev/dl](https://go.dev/dl/)).
- Terminal with ANSI / UTF-8 TrueColor support.
- Recommended terminal window resolution: **64 columns × 26 rows** (or larger).

### 1. Clone the repository
```bash
git clone https://github.com/mvsoares/tetris.git
cd tetris
```

### 2. Build the binaries
Using `make`:
```bash
make build
```
Or using the Go CLI directly:
```bash
go build -o tetris ./cmd/tetris
go build -o train ./cmd/train
go build -o analyze ./cmd/analyze
go build -o analyze_losses ./cmd/analyze_losses
go build -o learn ./cmd/learn
```

Or run directly without building:
```bash
go run ./cmd/tetris
```

---

## 🚀 Usage Guide

### 1. Play in the Terminal (Interactive Mode)
Launch the graphical terminal UI:
```bash
./tetris
```
*Tip*: Press `M` to choose between standard heuristic AI (v2), 10-piece lookahead, or hybrid risk model. Press `B` or `Tab` to toggle Auto-Play. Press `L` to switch language at any time.

### 2. Run Headless Simulations (`./train`)
Generate hundreds of thousands of moves at high speed for machine learning training or survival benchmarking:
```bash
# Simulate 100 games across 15 concurrent workers with a 100 MB write buffer:
./train -games 100 -workers 15 -file logs/plays.jsonl

# Truncate existing log before starting:
./train -clean -games 50 -workers 10

# Continuous mode (runs until Ctrl+C):
./train -games 0 -workers 50

# Reproducible benchmark: exactly 1,000 games, up to 1,000 pieces per game:
go run ./cmd/train -quiet -games 1000 -workers 16 -seed 1 -max-moves 1000 -file logs/benchmark-placement.jsonl

# Same seeds, using actual gravity-governed AI movements:
go run ./cmd/train -quiet -mode gameplay -games 1000 -workers 16 -seed 1 -max-moves 1000 -file logs/benchmark-gameplay.jsonl
```

The game seed for index `n` is computed as `seed + n` regardless of worker count.
- The default `placement` mode evaluates direct drop positions without path execution.
- The `gameplay` mode steps `StepAI` every virtual 55 ms and ticks gravity according to level speeds without real-time delays.

### 3. Experimental 10-Piece Lookahead
In-game, press **M** to open the AI menu. Select **Lookahead 10 pieces (beam 4)** using arrow keys or numbers and confirm with `Enter`.
From the command line:
```bash
go run ./cmd/tetris -lookahead 10 -beam-width 4
go run ./cmd/train -quiet -mode gameplay -lookahead 10 -beam-width 4 -games 20 -workers 4 -seed 101 -max-moves 1000 -file logs/beam-pilot-new.jsonl
```

### 4. Learned Risk Probability Model
`./tetris` loads `models/move-risk.json` at startup. Launch with:
```bash
./tetris -learned -model models/move-risk.json
./train -quiet -mode gameplay -learned -model models/move-risk.json -games 20 -workers 4 -seed 5001 -max-moves 1000 -file logs/hybrid-new.jsonl
```
The status panel displays `P50 sim`: estimated survival probability over 50 simulated future placements based on 7-bag rollouts.

To run offline training with simulated alternatives:
```bash
# Collect gameplay decisions with full state transitions:
./train -quiet -mode gameplay -games 30 -workers 4 -seed 1001 -max-moves 1000 -buf-kb 4096 -file logs/learning-source-new.jsonl

# Simulate alternatives and atomically export an updated model:
./learn -file logs/learning-source-new.jsonl -out models/move-risk-new.json -decisions 120 -candidates 4 -rollouts 4 -horizon 50 -workers 4 -seed 20260927
```

### 5. Analyze Gameplay Datasets (`./analyze`)
Generate comprehensive aggregate statistics from any `.jsonl` log file:
```bash
./analyze -file logs/plays.jsonl
```

### 6. Forensic Loss Diagnosis (`./analyze_losses`)
Diagnose root causes of game overs (top-out piece, 10-column profile, hole counts, defense mode):
```bash
./analyze_losses -file logs/plays.jsonl

# Structured JSON output with percentiles and termination reasons:
go run ./cmd/analyze_losses -file logs/benchmark-gameplay.jsonl -json
```

---

## 🧠 How the Auto-Play AI Algorithm Works

The Auto-Play agent combines constrained legal pathfinding, future piece projection, and an adaptive heuristic function that balances long-term survival with high-scoring Tetrises.

### 1. Decision Pipeline Each Turn (Policy v2)
1. **Threat Assessment**: Evaluates board hazard state, contours, and identifies legal navigation paths from the spawn position using lateral shifts, dual-direction SRS wall kicks, and inter-action gravity.
2. **Placement Evaluation**: Simulates lock and line clearing on every valid terminal candidate, scoring height, holes, surface roughness, spawn chimney headroom, structural support, and well openness.
3. **Queue Lookahead**: Ranks candidates against the next piece in the queue, deepening lookahead during danger states.
4. **Hold Strategy**: Compares active candidates against Hold alternatives and selects the optimal path.
5. **Execution**: Emits the action sequence. If an unexpected blockage or desync occurs, re-plans immediately rather than executing a blind drop.

### 2. 9-0 Stacking & The 4 Golden Rules of the Well
The **9-0 stacking strategy** confines block placement to columns **0 through 8**, keeping **column 9 empty** to score 4-line Tetrises with vertical `I` pieces:
1. **Vertical `I` in the well**: Rewarded heavily for 4-line clears; partial clears or debris left in the well incur penalties.
2. **Column 9 overhangs**: Heavily penalized when column 9 exceeds the height of column 8.
3. **Well blockage**: Covering empty cells in column 9 receives severe penalties.
4. **Depth and lateral access**: Excessive well depth and neighboring spikes penalize the evaluation score.

During emergency cleanup, well protection penalties are reduced to **8%**, enabling sacrificial partial clears when maintaining an open well is too dangerous.

### 3. Specialized Handling for Pieces (O, S, Z)
- 🟨 **O-Piece (2×2 - Flat Base Preference)**:
  - Requires a flat 2-cell base of identical height ($\Delta h = 0$).
  - **Flat Platform Bonus**: Receives a +3,500 score bonus when placed on an even foundation.
  - **Step Penalty**: Penalizes placement over steps of height $\ge 2$ ($-3,500 \times \Delta h$) which would create holes or spikes.
  - **Preventive Hold**: On high bumpiness terrains without a 2-cell flat slot, v2 holds `O` when a flexible alternative exists.
- 🟩 **S-Piece & 🟥 Z-Piece (Horizontal Alignment & Hold)**:
  - **Horizontal Preference**: Horizontal orientations receive bonuses (+1,500 pts); vertical placements in uneven terrains are strongly penalized (-3,500 pts).
  - **Emergency Swap**: If `S` or `Z` arrives under board stress (height $\ge 9$, bumpiness $\ge 8$, or existing holes), v2 prioritizes swapping with the Hold queue.

### 4. Terrain Contour Management
- 🥣 **Bowl Profile vs. Central Dome**:
  - Prevents blocks from piling up in the center (columns 3–5). Keeps terrain flat or slightly concave, ensuring clear spawn and rotation clearance at the top.
- 🏞️ **Left Cliff Suppression (Column 0)**:
  - Prevents column 0 from dropping 2+ cells below column 1, eliminating single-cell blind ravines where pieces would get stuck.

### 5. Dynamic Reactive Cleanup Mode
The engine dynamically calculates its defense threshold via `GetDynamicCleanupThreshold`:
- Base threshold starts at 14, decreases by 3 per existing hole, and decreases further if central height $\ge 9$ or bumpiness $\ge 8$.
- **Early Activation**:
  - If a hole forms at height $\ge 8$, cleanup mode triggers immediately to uncover it before new blocks bury it.
  - If bumpiness reaches $\ge 13$ at height $\ge 8$, cleanup mode activates to smooth terrain.
- **Drastic Weight Shifts in Cleanup Mode**:
  - Penalty for creating new holes jumps from 15,000 to **55,000 points**.
  - Partial line clears (1, 2, 3 lines) receive major positive bonuses (+35,000 to +95,000 pts) to rapidly deflate stack height.

### 📊 Benchmark Performance Evolution

Reproducible benchmark comparisons capped at 1,000 pieces per game:

| Experiment | Previous Policy | New Policy |
|---|---:|---:|
| Survival — 1,000 seeds (v1 → v2) | 41.1% | 88.6% |
| Mean Score — same 1,000 seeds | 879,534 | 1,313,620 |
| Tetris Share — same 1,000 seeds | 59.9% | 54.5% |
| Survival — 20-seed pilot (v2 → 10-piece lookahead) | 18/20 | 20/20 |
| Mean Score — same 20 seeds | 1,343,355 | 1,496,056 |
| Tetris Share — same 20 seeds | 55.6% | 70.7% |

*The v2 policy achieved zero plan misses across 951,555 planned placements.*

---

## 🏛️ Codebase Architecture

```
tetris/
├── cmd/
│   ├── tetris/           # Interactive terminal game (Bubble Tea UI)
│   ├── train/            # Headless concurrent CLI simulation orchestrator
│   ├── analyze/          # Analytical CLI reporting aggregate metrics
│   ├── analyze_losses/   # Forensic diagnosis of top-outs and game over causes
│   └── learn/            # Offline model training with simulated counterfactuals
├── internal/
│   ├── engine/           # Pure Tetris domain logic (100% decoupled from UI)
│   │   ├── board.go      # BitBoard 20xuint16, collisions, line clear, bitwise masks
│   │   ├── piece.go      # Definitions and SRS rotations for 7 tetrominos (I, J, L, O, S, T, Z)
│   │   ├── randomizer.go # Fair 7-Bag generation system
│   │   ├── game.go       # Core game loop, scoring rules, levels, and state machine
│   │   ├── ai.go         # Heuristics, adaptive lookahead, and dynamic hazard detection
│   │   ├── path.go       # Legal gravity-aware pathfinding and replanning
│   │   ├── beam.go       # Multi-ply lookahead beam search (up to 10 placements)
│   │   ├── learned.go    # Local risk model, inference, and heuristic fallback
│   │   └── learned_train.go # Counterfactual rollout simulations and training
│   ├── i18n/             # Full internationalization dictionary (9 languages)
│   ├── logger/           # Non-blocking async logger with 100 MB buffered channel
│   ├── simulator/        # Concurrent worker pool for parallel simulations
│   ├── ui/               # Terminal presentation layer using Lipgloss and Bubble Tea
│   └── version/          # Centralized version control (v1.0.0) and build metadata
├── packaging/            # nfpm packaging definitions (.deb, .rpm)
├── scripts/              # Build, cross-compilation, packaging, and install scripts
├── Formula/              # Homebrew tap formula for macOS and Linux
├── models/               # Versioned machine learning model weights (move-risk.json)
└── logs/                 # Output directory for JSON Lines datasets (.jsonl)
```

---

## ⚙️ Command Line Options

```text
Usage: tetris [flags]

Flags:
  -v, --version        Print Tetris version (v1.0.0) and exit
  --lang string        Interface language (pt-br, en, es, fr, it, de, ru, ja, zh)
  --idiom string       Alias for --lang
  --learned            Enable experimental hybrid AI with trained risk model
  --lookahead int      Experimental lookahead depth: 0 (v2 default) or 1..10
  --beam-width int     Beam width for lookahead search (1..64, default: 4)
  --model string       Path to custom risk model (default: models/move-risk.json)
  --train              Run headless simulations in background
  --games int          Total games to run in headless mode (default: 100)
  --workers int        Concurrent workers pool size (1 to 50)
  --file string        Output path for JSONL dataset log
```

---

## 🧪 Automated Testing Suite

The codebase features comprehensive unit test coverage across all domain packages:

```bash
# Run all unit tests:
go test ./...

# Run race condition detector:
go test -race ./...

# Run static analysis and linting:
go vet ./...
```

---

## 📄 License

This project is open-source software licensed under the [MIT License](LICENSE).
