# 🎮 Tetris CLI em Go

Uma implementação moderna, fluida e de alto desempenho do clássico **Tetris** para terminal, desenvolvida em **Go** com a arquitetura [Bubble Tea](https://github.com/charmbracelet/bubbletea) (The Elm Architecture) e estilizada com [Lipgloss](https://github.com/charmbracelet/lipgloss).

Inclui um **Auto-Play com IA Heurística Adaptativa**, um **motor de simulação headless concorrente** capaz de processar mais de 4.000 jogadas/segundo, um **logger assíncrono com buffer de 100 MB** para geração de datasets JSONL e ferramentas CLI de análise profunda de partidas e derrotas.

---

## 🌟 Principais Recursos

- 🎨 **Interface Rica no Terminal**: Renderização TrueColor sem cintilação (*flicker-free*) no buffer alternativo (`tea.WithAltScreen`), com proporção de blocos 1:2 (`██`) e proteção de redimensionamento (`SIGWINCH`).
- 🕹️ **Mecânicas Modernas Oficiais**:
  - Randomizador **7-Bag** justo (sem secas prolongadas de peças).
  - Rotação com **Wall Kicks** (evita travamentos em paredes e bordas).
  - **Peça Fantasma** (*Ghost Piece*) semitransparente (`░░`).
  - **Hold Queue** com limite de 1 troca por travamento.
  - **Contador de Tetrises** na interface (painel lateral e tela de Game Over).
  - **Gravidade Progressiva**: A velocidade de queda aumenta a cada 10 linhas limpas.
- 🤖 **Auto-Play Inteligente (IA com Stacking 9-0 & Lookahead)**:
  - **Estratégia 9-0**: Constrói a pilha nas colunas 0 a 8 e mantém a coluna 9 aberta para a chegada da peça `I`.
  - **Lookahead de 2 camadas**: Avalia a peça atual e a próxima peça para evitar bloqueios de relevo.
  - **Heurística Dinâmica de Perigo**: Substitui limiares rígidos por análise contínua de terreno, buracos e proteção do corredor de spawn (colunas 3 a 6).
- ⚡ **Simulador Headless em Background (`./train`)**:
  - Worker pool concorrente com até **20 partidas paralelas** usando todas as threads da CPU.
  - Throughput de **> 4.000 jogadas/segundo**.
- 💾 **Logger Assíncrono com Buffer de 100 MB**:
  - Buffer de gravação de 100 MB via `bufio.Writer` e canal em fila de 131.072 itens.
  - Grava datasets completos em JSON Lines (`.jsonl`) com matriz `20x10`, relevo, buracos e pontuação sem penalizar o framerate.
- 📊 **Ferramentas CLI Analíticas**:
  - `./analyze`: Estatísticas agregadas, taxa de Tetris, pontuação e distribuição de linhas.
  - `./analyze_losses`: Diagnóstico profundo da causa raiz de cada Game Over (top-outs, relevo e buracos).

---

## 🕹️ Controles do Jogo

```text
┌──────────────────────┬────────────────────────────────────────────────────────┐
│ Tecla                │ Ação / Função                                          │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ ← / A / H            │ Mover peça para a Esquerda                             │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ → / D / L            │ Mover peça para a Direita                              │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ ↓ / S / J            │ Soft Drop (queda acelerada com bônus de pontuação)     │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ Espaço               │ Hard Drop (queda instantânea e travamento)             │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ ↑ / W / X            │ Rotacionar no sentido Horário                          │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ Z                    │ Rotacionar no sentido Anti-horário                     │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ C / H                │ Guardar / Trocar peça no Hold                          │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ B / Tab              │ Ativar / Desativar Auto-Play (IA joga automaticamente) │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ 4 / I                │ Forçar chegada de 4 peças de Linha (I) consecutivas    │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ T                    │ Setup Instantâneo de 4 Linhas para Tetris              │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ P                    │ Pausar / Retomar partida                               │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ R                    │ Reiniciar jogo                                         │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ Q / Esc / Ctrl+C     │ Sair do jogo                                           │
└──────────────────────┴────────────────────────────────────────────────────────┘
```

---

## 📥 Instalação e Compilação

### Pré-requisitos
- **Go 1.20** ou superior instalado ([golang.org](https://go.dev/dl/)).
- Terminal com suporte a cores ANSI / UTF-8.
- Resolução recomendada de terminal: **64 colunas × 26 linhas** (ou maior).

### 1. Clonar o repositório
```bash
git clone https://github.com/mvsoares/tetris.git
cd tetris
```

### 2. Compilar todos os binários
```bash
# Compila os utilitários do projeto
go build -o tetris ./cmd/tetris
go build -o train ./cmd/train
go build -o analyze ./cmd/analyze
go build -o analyze_losses ./cmd/analyze_losses
```

Ou execute diretamente com `go run`:
```bash
go run cmd/tetris/main.go
```

---

## 🚀 Como Usar

### 1. Jogar no Terminal (Modo Interativo)
Inicie a interface gráfica no seu terminal:
```bash
./tetris
```
*Dica*: Pressione `B` a qualquer momento durante a partida para ver a IA jogando ao vivo com 9-0 stacking e limpando 4 linhas consecutivas!

### 2. Simular Partidas em Background (`./train`)
Gere centenas de milhares de jogadas em alta velocidade para datasets de aprendizado de máquina ou benchmarks de sobrevivência:
```bash
# Simular 100 partidas com 15 workers simultâneos e buffer de 100 MB:
./train -games 100 -workers 15 -file logs/plays.jsonl

# Limpar o log existente antes de iniciar:
./train -clean -games 50 -workers 10

# Modo contínuo (roda sem parar até pressionar Ctrl+C):
./train -games 0 -workers 20
```

### 3. Analisar o Dataset de Jogadas (`./analyze`)
Gera um relatório estatístico completo das partidas registradas:
```bash
./analyze -file logs/plays.jsonl
```

### 4. Diagnóstico de Derrotas (`./analyze_losses`)
Analisa a causa raiz de cada Game Over (qual peça causou o topo, perfil de altura das 10 colunas, número de buracos no momento da morte e se a IA estava em modo defensivo):
```bash
./analyze_losses -file logs/plays.jsonl
```

---

## 🧠 Heurística da IA & Evolução do Algoritmo

O motor de IA passou por uma evolução profunda através da análise de mais de **500.000 jogadas reais**.

### O Problema do Limiar Rígido de 65%
Anteriormente, a IA utilizava uma regra binária:
- Se a altura fosse `< 13 linhas` (65% do tabuleiro), ela penalizava limpezas simples com `-3.000 pts` e recusava limpar linhas, focando exclusivamente em Tetris.
- Se houvesse uma peça `I` no Hold, ela desativava o modo de emergência.

**Resultado dos logs de perda**: 60.7% das mortes aconteciam no modo normal! A IA acumulava buracos na altura 10 ou 11, formava um pico central nas colunas 3 a 5 (onde as peças nascem) e sofria colisão imediata ao spawnar peças `O`, `Z` ou `S`.

### A Solução: Heurística Dinâmica Adaptativa
Implementamos uma abordagem orientada a risco contínuo:
1. **`GetDynamicCleanupThreshold`**: O limiar varia de 8 a 15 linhas dependendo da saúde do tabuleiro. Tabuleiros planos sem buracos permitem construir com segurança até 15 linhas; tabuleiros com buracos ativam a defesa imediatamente (cada buraco reduz o limiar em 2 linhas).
2. **Proteção do Corredor de Spawn (`spawnChutePenalty`)**: Penaliza fortemente o acúmulo de blocos nas colunas 3 a 6 em relação às laterais, garantindo que o relevo seja plano ou em formato de bacia (*bowl*), permitindo que peças recém-nascidas deslizem para as laterais sem colidir no teto.
3. **Valoração Flexível de Linhas**: Quando o terreno se eleva, limpezas de 1, 2 e 3 linhas passam a receber recompensas positivas (`+2.000` a `+6.000`), estabilizando o tabuleiro sem sacrificar a sobrevivência.
4. **Resgate pelo Hold**: Em alturas críticas (>= 14 linhas), o modo defensivo é forçado mesmo que haja uma peça `I` no Hold.

### Comparativo de Desempenho

```text
┌──────────────────────────────────────┬──────────────────────┬──────────────────────┬─────────────┐
│ Métrica de Desempenho                │ Heurística Anterior  │ Heurística Dinâmica  │ Evolução    │
├──────────────────────────────────────┼──────────────────────┼──────────────────────┼─────────────┤
│ Mediana de Jogadas por Partida       │ 2.215 jogadas        │ 3.900 jogadas        │ +76.1%      │
├──────────────────────────────────────┼──────────────────────┼──────────────────────┼─────────────┤
│ Média de Jogadas por Partida         │ 3.186 jogadas        │ 4.676 jogadas        │ +46.8%      │
├──────────────────────────────────────┼──────────────────────┼──────────────────────┼─────────────┤
│ Mediana de Linhas Limpas por Partida │ 874 linhas           │ 1.544 linhas         │ +76.7%      │
├──────────────────────────────────────┼──────────────────────┼──────────────────────┼─────────────┤
│ Mortes Cegas em Modo Normal (< 65%)  │ 60.7% das derrotas   │ 14.1% das derrotas   │ -76.8%      │
├──────────────────────────────────────┼──────────────────────┼──────────────────────┼─────────────┤
│ Throughput de Gravação (Buffer 100M) │ ~1.200 jogadas/s     │ > 4.000 jogadas/s    │ +233%       │
└──────────────────────────────────────┴──────────────────────┴──────────────────────┴─────────────┘
```

---

## 🏛️ Estrutura do Código

```
tetris/
├── cmd/
│   ├── tetris/           # Jogo interativo no terminal (Bubble Tea UI)
│   ├── train/            # Orquestrador CLI headless de simulação concorrente
│   ├── analyze/          # Relatório analítico de métricas agregadas
│   └── analyze_losses/   # Diagnóstico profundo de top-outs e causa raiz de mortes
├── internal/
│   ├── engine/           # Lógica pura do Tetris (100% desacoplada da UI)
│   │   ├── board.go      # Grid 10x20, colisões, line clearing e wall kicks
│   │   ├── piece.go      # Definição e rotações dos 7 tetrominós (I, J, L, O, S, T, Z)
│   │   ├── randomizer.go # Sistema de geração 7-Bag
│   │   ├── game.go       # Loop do jogo, pontuação, níveis e estados
│   │   └── ai.go         # Heurísticas de avaliação, lookahead 2-ply e perigo dinâmico
│   ├── logger/           # Gravador assíncrono com canal e buffer bufio de 100 MB
│   ├── simulator/        # Pool de workers concorrentes para partidas em paralelo
│   └── ui/               # Renderização de terminal com Lipgloss e Bubble Tea
└── logs/                 # Destino dos datasets em JSON Lines (.jsonl)
```

---

## 🧪 Testes Automatizados

O projeto possui cobertura abrangente de testes unitários em todos os pacotes do domínio:

```bash
go test -v ./...
```

---

## 📄 Licença

Este projeto é disponibilizado sob a licença [MIT](LICENSE).
