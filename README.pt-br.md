# 🎮 Tetris CLI em Go

[![Release](https://img.shields.io/badge/release-v1.0.0-blue.svg)](https://github.com/mvsoares/tetris/releases)
[![Go Version](https://img.shields.io/badge/go-1.26+-00ADD8.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![i18n](https://img.shields.io/badge/languages-9%20Idiomas%20(PT--BR%20%7C%20EN%20%7C%20ES%20%7C%20FR%20%7C%20IT%20%7C%20DE%20%7C%20RU%20%7C%20JA%20%7C%20ZH)-orange.svg)](#-idiomas--multilingual-support)

*Traduções: [🇺🇸 English](README.md) | [🇧🇷 Português](README.pt-br.md)*

Uma implementação moderna, fluida e de alto desempenho do clássico **Tetris** para terminal, desenvolvida em **Go** com a arquitetura [Bubble Tea](https://github.com/charmbracelet/bubbletea) (The Elm Architecture) e estilizada com [Lipgloss](https://github.com/charmbracelet/lipgloss).

Inclui **suporte completo a 9 idiomas** (Português, Inglês, Espanhol, Francês, Italiano, Alemão, Russo, Japonês e Chinês Simplificado), **Auto-Play com IA Heurística Adaptativa e Motor BitBoard**, **motor de simulação headless concorrente**, **logger assíncrono com buffer configurável** para geração de datasets JSONL e ferramentas CLI de análise profunda de partidas e derrotas.

Veja o [histórico de alterações](CHANGELOG.md) e os [resultados dos benchmarks](benchmarks/README.md).

---

## 🌐 Idiomas / Multilingual Support

O Tetris conta com tradução nativa completa para **9 idiomas**:
- 🇧🇷 **Português do Brasil** (`pt-br`, `pt`) — *padrão*
- 🇺🇸 **English** (`en`, `english`)
- 🇪🇸 **Español** (`es`, `spanish`)
- 🇫🇷 **Français** (`fr`, `french`)
- 🇮🇹 **Italiano** (`it`, `ita`, `italian`)
- 🇩🇪 **Deutsch** (`de`, `german`, `deutsch`)
- 🇷🇺 **Русский** (`ru`, `russian`)
- 🇯🇵 **日本語** (`ja`, `japanese`)
- 🇨🇳 **简体中文** (`zh`, `chinese`)

### Como alterar o idioma:
1. **Durante o jogo**: Pressione a tecla **`L`** a qualquer momento para abrir o menu interativo de seleção de idioma (1 a 9 ou setas).
2. **Na linha de comando**: Inicie o jogo com a flag `--lang` ou `--idiom`:
   ```bash
   tetris --lang=ja        # 日本語 (Japonês)
   tetris --lang=ru        # Русский (Russo)
   tetris --lang=de        # Deutsch (Alemão)
   tetris --lang=zh        # 简体中文 (Chinês)
   tetris --lang=english   # English
   tetris --lang=spanish   # Español
   tetris --lang=french    # Français
   tetris --lang=ita       # Italiano
   tetris --lang=pt-br     # Português do Brasil
   ```

---

## 🌟 Principais Recursos

- 🌐 **Multilíngue Global (i18n)**: 9 idiomas suportados com alternador dinâmico em tempo de execução (`L`, teclas 1..9) e flags CLI (`--lang`, `--idiom`).
- 📦 **Empacotamento Oficial (v1.0.0)**: Pacotes prontos para **Debian/Ubuntu (`.deb`)**, **Fedora/RHEL (`.rpm`)**, **macOS (`.tar.gz` / Homebrew)** e **Windows (`.zip`)**.
- 🎨 **Interface Rica no Terminal**: Renderização TrueColor sem cintilação (*flicker-free*) no buffer alternativo (`tea.WithAltScreen`), com proporção de blocos 1:2 (`██`) e proteção de redimensionamento (`SIGWINCH`).
- 🕹️ **Mecânicas Modernas Oficiais**:
  - Randomizador **7-Bag** justo (sem secas prolongadas de peças).
  - Rotação com **Wall Kicks** (evita travamentos em paredes e bordas).
  - **Peça Fantasma** (*Ghost Piece*) semitransparente (`░░`).
  - **Hold Queue** com limite de 1 troca por travamento.
  - **Contador de Tetrises** na interface (painel lateral e tela de Game Over).
  - **Gravidade Progressiva**: A velocidade de queda aumenta a cada 10 linhas limpas.
- 🤖 **Auto-Play Inteligente (IA com Stacking 9-0, BitBoard & Lookahead)**:
  - **Motor BitBoard**: Execução ultrarrápida com representação bitwise de 40 bytes na stack.
  - **Estratégia 9-0**: Constrói a pilha nas colunas 0 a 8 e mantém a coluna 9 aberta para a chegada da peça `I`.
  - **Rotas legais com gravidade**: Planeja movimentos reais e recalcula a rota quando necessário.
  - **Lookahead adaptativo**: A política v2 avalia a próxima peça e aprofunda a busca quando há perigo.
  - **Lookahead experimental de 10 peças**: Busca em feixe com hold e remoção de estados duplicados, selecionável pelo menu `M`.
  - **Modelo local de risco**: Estimativas treinadas com alternativas simuladas; terceira opção experimental no menu. A v2 permanece padrão.
- ⚡ **Simulador Headless em Background (`./train`)**:
  - Worker pool concorrente com até **50 partidas paralelas** usando todas as threads da CPU.
  - Seeds reproduzíveis e modos `placement` e `gameplay`, com motivos explícitos de encerramento.
- 💾 **Logger Assíncrono com Buffer de 100 MB**:
  - Buffer de gravação de 100 MB via `bufio.Writer` e canal em fila de 131.072 itens.
  - Grava JSON Lines (`.jsonl`) com matrizes `20x10`, relevo, buracos, pontuação e telemetria de planos.
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
│ → / D / L (no jogo)  │ Mover peça para a Direita                              │
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
│ M                    │ Abrir menu para escolher a política de IA             │
├──────────────────────┼────────────────────────────────────────────────────────┤
│ L                    │ Abrir menu de seleção de Idioma (PT-BR, EN, ES, FR, IT)│
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

## 📦 Instalação dos Pacotes Oficiais (v1.0.0)

### Opção 1: Instalador Automático Universal (Linux e macOS)
Execute no seu terminal para detectar a sua distribuição/SO e instalar automaticamente a versão mais recente:
```bash
curl -fsSL https://raw.githubusercontent.com/mvsoares/tetris/main/scripts/install.sh | bash
```

### Opção 2: Debian / Ubuntu (`.deb`)
Baixe o pacote da release para sua arquitetura (`amd64` ou `arm64`) e instale:
```bash
# Para arquitetura x86_64 / amd64:
sudo dpkg -i tetris_1.0.0_amd64.deb

# Para arquitetura ARM64:
sudo dpkg -i tetris_1.0.0_arm64.deb
```

### Opção 3: Fedora / Red Hat / CentOS (`.rpm`)
```bash
# Para arquitetura x86_64:
sudo rpm -Uvh tetris-1.0.0-1.x86_64.rpm
# ou
sudo dnf install tetris-1.0.0-1.x86_64.rpm

# Para arquitetura aarch64 (ARM64):
sudo rpm -Uvh tetris-1.0.0-1.aarch64.rpm
```

### Opção 4: macOS (Apple Silicon ou Intel)
Baixe o arquivo compactado da release:
```bash
# Apple Silicon (M1/M2/M3/M4):
tar -xzf tetris_1.0.0_darwin_arm64.tar.gz
sudo mv tetris /usr/local/bin/

# Mac Intel:
tar -xzf tetris_1.0.0_darwin_amd64.tar.gz
sudo mv tetris /usr/local/bin/
```
Ou instale via [Homebrew](Formula/tetris.rb):
```bash
brew install mvsoares/tetris/tetris
```

---

## 📥 Instalação e Compilação

### Pré-requisitos

- **Go 1.26.6** ou superior instalado, conforme `go.mod` ([go.dev](https://go.dev/dl/)).
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
go build -o learn ./cmd/learn
```

Ou execute diretamente com `go run`:
```bash
go run ./cmd/tetris
```

---

## 🚀 Como Usar

### 1. Jogar no Terminal (Modo Interativo)
Inicie a interface gráfica no seu terminal:
```bash
./tetris
```
*Dica*: Pressione `M` para escolher entre a IA v2 e o lookahead de 10 peças; depois use `B` ou Tab para ativar o Auto-Play. Alterar a política reinicia a rodada. Após atualizar o código, recompile `./tetris`: executáveis locais não são versionados.

### 2. Simular Partidas em Background (`./train`)
Gere centenas de milhares de jogadas em alta velocidade para datasets de aprendizado de máquina ou benchmarks de sobrevivência:
```bash
# Simular 100 partidas com 15 workers simultâneos e buffer de 100 MB:
./train -games 100 -workers 15 -file logs/plays.jsonl

# Limpar o log existente antes de iniciar:
./train -clean -games 50 -workers 10

# Modo contínuo (roda sem parar até pressionar Ctrl+C):
./train -games 0 -workers 50

# Benchmark reproduzível: exatamente 1000 partidas, até 1000 peças por partida:
go run ./cmd/train -quiet -games 1000 -workers 16 -seed 1 -max-moves 1000 -file logs/benchmark-placement.jsonl

# Mesmas seeds, usando movimentos reais da IA e gravidade por nível:
go run ./cmd/train -quiet -mode gameplay -games 1000 -workers 16 -seed 1 -max-moves 1000 -file logs/benchmark-gameplay.jsonl
```

A seed da partida de índice `n` é `seed+n`, independentemente da quantidade de workers. O modo padrão `placement` avalia encaixes diretos, sem executar a sequência de movimentos. O modo `gameplay` executa `StepAI` a cada 55 ms virtuais e `Tick` no intervalo de gravidade do nível, sem esperar tempo real. Ele reproduz as regras e os intervalos nominais da interface; atrasos de renderização, processamento e ordem de mensagens no terminal podem produzir diferenças.

A política `heuristic-9-0-v2` busca caminhos legais a partir da posição atual, incluindo wall kicks, rotações nos dois sentidos e gravidade entre ações. Ao detectar desvio do caminho, calcula uma nova rota; não força um hard drop de watchdog. O poço continua reservado no modo normal, mas suas penalidades são reduzidas em emergências. Para comparar com penalidades rígidas também em emergência, use `-reserve-well` (registrado como `heuristic-9-0-v2-strict-well`). A busca usa até 2.048 estados e 32 ações; conserva o primeiro caminho para cada posição/rotação, podendo omitir caminhos mais longos com outra fase de gravidade. A próxima peça ainda usa avaliação aproximada de encaixes no lookahead.

`-max-moves` limita apenas peças colocadas, não linhas limpas. Partidas que atingem esse limite são sobreviventes censurados: não contam como derrotas, e sua duração não deve ser interpretada como tempo até perder.

### Lookahead experimental de 10 colocações

No jogo, pressione **M** para abrir o menu de IA. Escolha a política atual (v2),
**Lookahead 10 peças** ou **IA híbrida experimental** com ↑/↓ ou 1/2/3 e confirme com Enter. Esc/M cancela.
O jogo fica congelado enquanto o menu está aberto. Alterar a política reinicia
a partida, preservando o recorde e o estado do Auto-Play; confirmar a opção já
ativa não reinicia. Use B/Tab para ligar ou desligar o Auto-Play.

```bash
go run ./cmd/tetris -lookahead 10 -beam-width 4
go run ./cmd/train -quiet -mode gameplay -lookahead 10 -beam-width 4 -games 20 -workers 4 -seed 101 -max-moves 1000 -file logs/beam-pilot-new.jsonl
```

`-lookahead 0` mantém a política v2 padrão. O modo experimental amplia a fila conhecida para 10 peças futuras, também exibidas como letras no painel NEXT. A profundidade conta a peça atual entre as 10 colocações; a peça extra suporta o hold vazio. A sequência gerada não muda. A busca considera hold em cada camada, conserva as melhores continuações distintas até `beam-width` e executa somente a primeira colocação planejada, recalculando após cada peça. Ela não é exaustiva: a poda pode descartar a melhor sequência global, e as colocações futuras ainda usam o modelo aproximado sem rotas completas de gravidade/SRS. `search_depth` e `search_nodes` registram a profundidade efetivamente alcançada e o número de encaixes avaliados nos movimentos com plano.

Os resultados e limites estão no [teste de viabilidade](benchmarks/2026-09-26-lookahead.md). Larguras maiores custam mais CPU/memória e podem atrasar os timers da UI; o simulador usa tempo virtual e não mede esse atraso real. Por isso o modo permanece opt-in.

### Modelo aprendido de probabilidades

`./tetris` carrega `models/move-risk.json` uma vez na inicialização, mas não ativa
a política híbrida automaticamente. Execute a partir da pasta do projeto, abra
o menu **M** e escolha **3**, ou use:

```bash
./tetris -learned -model models/move-risk.json
./train -quiet -mode gameplay -learned -model models/move-risk.json -games 20 -workers 4 -seed 5001 -max-moves 1000 -file logs/hybrid-new.jsonl
```

O painel mostra `P50 sim`: estimativa de sobreviver por 50 colocações (incluindo
a candidata), supondo continuidade pela v2 em **encaixe direto** e futuros 7-bag
amostrados além da fila conhecida. Não é probabilidade de vitória nem garantia
para o modo interativo. O JSON de decisão registra estimativas também para
alternativas legais; elas não precisam somar 100%, pois são probabilidades de
resultado, não uma distribuição de escolha.

O primeiro modelo melhorou o Brier de validação (0,0215 contra 0,0317 para um
prior constante), mas o piloto de gameplay em novas seeds teve **16/20**
sobreviventes contra **19/20** na v2. É experimental e **não é uma melhoria de
gameplay comprovada**. Modelo ausente, inválido, estado fora dos intervalos de
treino ou falta de vantagem suficiente mantêm o fallback heurístico.

Treino reproduzível, sempre com arquivos novos:

```bash
# Coleta decisões com fila, hold, pose, gravidade e alternativas:
./train -quiet -mode gameplay -games 30 -workers 4 -seed 1001 -max-moves 1000 -buf-kb 4096 -file logs/learning-source-new.jsonl
# Simula alternativas e publica atomicamente um modelo novo:
./learn -file logs/learning-source-new.jsonl -out models/move-risk-new.json -decisions 120 -candidates 4 -rollouts 4 -horizon 50 -workers 4 -seed 20260927
```

O ajuste é offline; jogar não altera os pesos nem inicia treino automaticamente.
Logs antigos sem `decision` não bastam para esse comando. Consulte a
[documentação do modelo](models/README.md) e o [piloto de validação](benchmarks/2026-09-27-risk.md).

### 3. Analisar o Dataset de Jogadas (`./analyze`)
Gera um relatório estatístico completo das partidas registradas:
```bash
./analyze -file logs/plays.jsonl
```

### 4. Diagnóstico de Derrotas (`./analyze_losses`)
Analisa a causa raiz de cada Game Over (qual peça causou o topo, perfil de altura das 10 colunas, número de buracos no momento da morte e se a IA estava em modo defensivo):
```bash
./analyze_losses -file logs/plays.jsonl

# Estatísticas estruturadas, incluindo percentis e motivos de encerramento:
go run ./cmd/analyze_losses -file logs/benchmark-gameplay.jsonl -json
```

O diagnóstico considera derrota somente sessões com `end_reason` igual a `top_out` ou `no_legal_move`. Limites (`move_limit`), cancelamentos (`cancelled`), saídas (`closed`) e reinícios (`restarted`) são separados. Logs antigos ou sessões incompletas ficam como `unknown`; não é possível recuperar com segurança o motivo de encerramento desses registros. `failed_piece` registra a peça que não conseguiu nascer ou obter um encaixe válido, e não a última peça colocada. Contagens por peça descrevem associação, não provam causalidade.

Cada encerramento registra seed, versão da política, modo de execução e tabuleiro final. Cada jogada registra tabuleiros e métricas antes/depois, recompensa incluindo soft/hard drop e o modo de limpeza da decisão. Linhas ocultas usadas pela busca não geram pontos extras de hard drop. `had_ai_plan`, `ai_plan_matched`, `plan_misses` e `watchdog_drops` permitem medir falhas de execução separadamente da heurística. O analisador verifica sequência de jogadas, recompensas e totais da sessão; registros JSON inválidos e erros de gravação/flush são reportados.

Veja o [benchmark de 26/09/2026](benchmarks/2026-09-26.md), com 1.000 seeds comparadas nos modos de encaixe e gameplay, resultados e prioridades para melhorar o auto-play.

As melhorias foram implementadas na política v2: o [reteste com as mesmas 1.000 seeds](benchmarks/2026-09-26-v2.md) elevou a sobrevivência até 1.000 peças de 41,1% para 88,6%, com zero desvios entre as 951.555 colocações que tinham plano. A participação de Tetrises caiu de 59,9% para 54,5%, enquanto a pontuação média aumentou 49,4%.

---

## 🧠 Como Funciona o Algoritmo de Auto-Play (IA Heurística)

O Auto-Play combina busca limitada de rotas legais, projeção de peças futuras e uma função heurística que equilibra sobrevivência e Tetrises. A política padrão é `heuristic-9-0-v2`; o lookahead de 10 colocações é experimental. Notas heurísticas não são pontos reais do jogo nem garantias de sobrevivência.

---

### 1. Fluxo de Decisão a Cada Turno

Na política v2:

1. Avalia o perigo do tabuleiro e procura rotas a partir da posição real da peça, com movimentos laterais, rotações SRS nos dois sentidos e gravidade entre ações.
2. Simula o travamento e a limpeza de linhas em cada candidato; avalia altura, buracos, relevo, corredor de nascimento, suporte e integridade do poço.
3. Compara os melhores candidatos usando a próxima peça e uma camada adicional quando há perigo. Peças futuras usam um modelo de encaixe aproximado.
4. Compara alternativas de hold e guarda a coluna, a rotação, a altura de aterrissagem, as ações e as posições esperadas.
5. Executa a rota; se houver desvio ou bloqueio, recalcula em vez de forçar uma queda cega.

A busca da rota atual é limitada a 2.048 estados e 32 ações. Conservar a primeira rota por posição/rotação pode excluir uma rota mais longa com outra fase de gravidade.

No modo de 10 colocações, uma busca em feixe substitui o ranqueamento adaptativo. Ela considera hold em cada camada, soma notas com desconto de 0,85 por profundidade e conserva até quatro estados distintos por padrão. Cores não fazem parte da identidade do estado; ocupação, peça atual, hold e posição na fila fazem. Apenas a primeira colocação é executada antes de uma nova busca.

---

### 2. A Estratégia 9-0 Stacking & As 4 Regras de Ouro do Poço

O empilhamento **9-0** privilegia as colunas **0 a 8** e reserva a **coluna 9** para Tetrises com a peça `I`. Essa preferência é expressa por penalidades, não por proibições absolutas:

1. **I vertical no poço**: limpar quatro linhas recebe forte recompensa; limpezas parciais ou blocos residuais têm penalidades.
2. **Espigões na coluna 9**: a nota cai quando a coluna 9 supera a coluna 8.
3. **Fechamento do poço**: cobrir vazios na coluna 9 recebe uma penalidade alta.
4. **Profundidade e acesso lateral**: poços excessivamente profundos e espigões próximos reduzem a nota.

Durante a limpeza de emergência, as penalidades do poço são reduzidas para **8%**, permitindo limpezas parciais quando preservar o poço seria arriscado. Penalidades por buracos e bloqueios continuam ativas. `-reserve-well` mantém as penalidades rígidas para comparação. Valores heurísticos não são pontos concedidos ao jogador.

---

### 3. Tratamento Especializado por Tipo de Peça (O, S, Z)

A IA possui preferências específicas para `O`, `S` e `Z`. Os benchmarks atuais não demonstram que essas peças sejam as principais causas de top-out: a última peça colocada não é necessariamente a que falhou ao nascer.

- 🟨 **Peça O (2×2 - Plataforma Plana)**:
  - Como a peça `O` não rotaciona e possui largura 2, ela exige uma base plana de duas células adjacentes de mesma altura ($\Delta h = 0$).
  - **Bônus de Plataforma Nivelada**: Recebe bônus heurístico de 3.500 quando há suporte nivelado, medido no tabuleiro **antes** da colocação.
  - **Penalidade de Degrau**: Penaliza severamente ($-3.500\text{ pts} \times \Delta h$) o encaixe sobre degraus de altura $\ge 2$, que deixariam a peça suspensa criando buracos ou espigões.
  - **Hold Preventivo para `O`**: Em terrenos com alta irregularidade ($\ge 12$) sem superfícies planas de 2 células, a política v2 favorece guardar `O` no hold quando uma alternativa flexível tem nota suficientemente próxima.
- 🟩 **Peça S e 🟥 Peça Z (Alinhamento Horizontal & Hold)**:
  - Peças diagonais causam problemas graves quando posicionadas verticalmente em espaços apertados, pois sua cauda inferior fica suspensa sobre vazios.
  - **Preferência Horizontal**: Posições horizontais recebem bônus ($+1.500\text{ pts}$), enquanto orientações verticais em terrenos irregulares são fortemente desencorajadas ($-3.500\text{ pts}$).
  - **Swap de Emergência no Hold**: Se uma peça `S` ou `Z` chega com o tabuleiro sob estresse (altura $\ge 9$, irregularidade $\ge 8$ ou presença de buracos), a política v2 favorece uma troca preventiva quando a alternativa de hold tem nota suficientemente próxima.

---

### 4. Supressão de Deformidades de Terreno

O relevo do tabuleiro é monitorado continuamente para prevenir a formação de armadilhas estruturais:

- 🥣 **Perfil em Bacia (*Bowl Profile*) vs. Domo Central**:
  - *Problema anterior*: O centro (colunas 3 a 5) acumulava blocos formando uma corcunda mais alta que as laterais, fazendo peças recém-nascidas colidirem logo na linha 0.
  - *Solução*: As colunas centrais 3 a 5 são penalizadas se superarem a altura das laterais (colunas 0–2 ou 7–8). O terreno é mantido plano ou ligeiramente côncavo, oferecendo espaço livre no topo para rotação e transladação imediata ao spawnar.
- 🏞️ **Supressão do Desfiladeiro da Esquerda (Coluna 0)**:
  - Evita que a coluna 0 fique 2 ou mais blocos mais baixa que a coluna 1, o que criaria um poço cego de 1 célula na borda esquerda onde a maioria das peças não conseguiria encaixar sem travar.

---

### 5. Modo de Limpeza Dinâmico e Reativo (Cleanup Mode)

Em vez de depender de uma altura arbitrária fixa (como o antigo limiar estático de 65%), a IA calcula em tempo real o perigo através de `GetDynamicCleanupThreshold`:

O limiar parte de 14, perde três unidades por buraco e, quando a altura central chega a 9, perde também `alturaCentral - 8`. Se a irregularidade exceder 8, subtrai `bumpiness - 7`. O resultado é limitado ao intervalo de 7 a 15; regras de emergência podem ativar a limpeza antes desse limiar.

- **Ativação Precoce**:
  - Se surgir o **primeiro buraco** na altura $\ge 8$, o modo de limpeza é ativado imediatamente para desenterrá-lo antes que ele seja coberto por novos blocos.
  - Se a irregularidade (*bumpiness*) atingir $\ge 13$ na altura $\ge 8$, o modo de limpeza atua para aplanar espigões antes que uma peça desfavorável cause *top-out*.
  - Se a altura central atingir $\ge 14$ ou a altura geral $\ge 16$, ativação de emergência máxima.
- **Diferenciação de Peça de Linha Imediata**:
  - Em terreno suficientemente seguro, uma peça `I` disponível na mão ou no hold pode adiar a limpeza; isso não verifica nem garante uma limpeza imediata de quatro linhas.
  - Esperas por peças futuras da fila são proibidas se o centro estiver em risco ($\ge 11$ linhas).
- **Mudança Drástica de Pesos no Modo Limpeza**:
  - O custo de criar um buraco sobe de $15.000$ para **$55.000\text{ pts}$**, desencorajando limpezas que deixem buracos; isso não garante sua ausência.
  - Linhas parciais (1, 2 e 3 linhas) passam a receber grandes recompensas positivas ($+35.000$ a $+95.000\text{ pts}$) para baixar a altura rapidamente.

---

### 6. Mecânica de Lock Delay para Jogadores Humanos

Em partidas manuais (`!AutoPlay`), o motor oferece tolerância antes do travamento:

- Quando a peça toca o chão ou a superfície dos blocos, ela não trava imediatamente: uma tolerância de **2 ticks de gravidade** é concedida. Sua duração varia com o nível; não é um atraso fixo de 500 ms.
- Cada movimento horizontal ou rotação bem-sucedida reinicia o temporizador (com limite de segurança de até 15 reinícios).
- Pressionar **Espaço (Hard Drop)** trava a peça instantaneamente sem atraso.
- No modo Auto-Play, a IA desativa o lock delay para manter execução de máxima velocidade.

---

### 📊 Evolução Comparativa de Desempenho

Comparações reproduzíveis com limite de 1.000 peças por partida:

| Experimento | Política anterior | Política nova |
|---|---:|---:|
| Sobrevivência — 1.000 seeds (v1 → v2) | 41,1% | 88,6% |
| Pontuação média — mesmas 1.000 seeds | 879.534 | 1.313.620 |
| Participação de Tetrises — mesmas 1.000 seeds | 59,9% | 54,5% |
| Sobrevivência — piloto de 20 seeds (v2 → lookahead 10) | 18/20 | 20/20 |
| Pontuação média — mesmas 20 seeds | 1.343.355 | 1.496.056 |
| Participação de Tetrises — mesmas 20 seeds | 55,6% | 70,7% |

A política v2 teve zero desvios nas 951.555 colocações com plano. O piloto de lookahead teve zero desvios nas 20.000 colocações, mas **não substitui uma validação com 1.000 partidas**. Chegar ao limite não significa sobrevivência indefinida; o piloto usa uma prévia maior e foi reutilizado durante o refinamento.

O lookahead de profundidade 10 e feixe 4 levou, em média, cerca de 22 ms por decisão no tabuleiro vazio da máquina testada, contra 1,77 ms na v2. Isso não é um limite de latência nem uma promessa de desempenho em outros computadores. Consulte os [relatórios e comandos completos](benchmarks/README.md).

---

## 🏛️ Estrutura do Código

```
tetris/
├── cmd/
│   ├── tetris/           # Jogo interativo no terminal (Bubble Tea UI)
│   ├── train/            # Orquestrador CLI headless de simulação concorrente
│   ├── analyze/          # Relatório analítico de métricas agregadas
│   ├── analyze_losses/   # Diagnóstico profundo de top-outs e causa raiz de mortes
│   └── learn/            # Treino offline com alternativas simuladas e validação por seed
├── internal/
│   ├── engine/           # Lógica pura do Tetris (100% desacoplada da UI)
│   │   ├── board.go      # BitBoard 20xuint16, colisões, line clearing e bitwise masks
│   │   ├── piece.go      # Definição e rotações dos 7 tetrominós (I, J, L, O, S, T, Z)
│   │   ├── randomizer.go # Sistema de geração 7-Bag
│   │   ├── game.go       # Loop do jogo, pontuação, níveis e estados
│   │   ├── ai.go         # Heurísticas, lookahead adaptativo e perigo dinâmico
│   │   ├── path.go       # Rotas legais com gravidade e replanejamento
│   │   ├── beam.go       # Lookahead experimental de até 10 colocações
│   │   ├── learned.go    # Modelo local, probabilidades e fallback
│   │   └── learned_train.go # Simulações contrafactuais e ajuste offline
│   ├── i18n/             # Internacionalização completa (EN, ES, PT-BR, FR, IT)
│   ├── logger/           # Gravador assíncrono com canal e buffer bufio de 100 MB
│   ├── simulator/        # Pool de workers concorrentes para partidas em paralelo
│   ├── ui/               # Renderização de terminal com Lipgloss e Bubble Tea
│   └── version/          # Controle centralizado de versão (v1.0.0) e build info
├── packaging/            # Configurações de empacotamento nfpm (.deb, .rpm)
├── scripts/              # Scripts de release, build multi-arch e instalador
├── Formula/              # Fórmula Homebrew para macOS e Linux
├── models/               # Modelo JSON versionado (move-risk.json)
└── logs/                 # Destino dos datasets em JSON Lines (.jsonl)
```

---

## ⚙️ Flags de Linha de Comando

```bash
tetris [flags]

Flags:
  -v, --version        Exibe a versão do Tetris (v1.0.0) e encerra
  --lang string        Idioma da interface (english, spanish, pt-br, french, ita)
  --idiom string       Alias para --lang
  --learned            Ativar IA híbrida experimental com modelo treinado
  --lookahead int      Lookahead experimental de colocações: 0 (v2) ou 1..10
  --beam-width int     Largura do feixe no lookahead (1..64, padrão: 4)
  --model string       Caminho para modelo de risco personalizado
  --train              Executa simulações em background
  --games int          Total de partidas no modo headless (padrão: 100)
  --workers int        Número de workers paralelos (1 a 50)
  --file string        Caminho do arquivo de log (.jsonl)
```

---

## 🧪 Testes Automatizados

O projeto possui cobertura abrangente de testes unitários em todos os pacotes do domínio:

```bash
go test ./...
go test -race ./...
go vet ./...
```

---

## 📄 Licença

Este projeto é disponibilizado sob a licença [MIT](LICENSE).
