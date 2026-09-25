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
  - Worker pool concorrente com até **50 partidas paralelas** usando todas as threads da CPU.
  - Throughput de **> 9.000 jogadas/segundo**.
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
./train -games 0 -workers 50
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

## 🧠 Como Funciona o Algoritmo de Auto-Play (IA Heurística)

O motor de Auto-Play foi desenvolvido com base na análise empírica de **mais de 2.000.000 de jogadas** simuladas concorrentemente. Ele combina **busca exaustiva de encaixes**, **verificação de acessibilidade física**, **projeção lookahead de 1 e 2 camadas (1-ply e 2-ply)** e uma **função de custo multi-critério** que equilibra sobrevivência em alta velocidade e pontuação máxima via *Tetrises* de 4 linhas.

---

### 1. Fluxo de Decisão a Cada Turno

Para cada turno de jogo, o algoritmo executa o seguinte pipeline:

```text
[Peça Atual + Fila Next + Hold]
              │
              ▼
 1. Geração de Candidatos (Rotações 0..3 × Colunas -3..9)
              │
              ▼
 2. Filtro de Acessibilidade Física (isReachable: traversals no topo sem colisão com espigões)
              │
              ▼
 3. Projeção de Queda Livre (GetGhostY: calcula posição exata de aterrissagem)
              │
              ▼
 4. Avaliação Heurística de 1º Nível (evaluatePlacement)
              │
              ▼
 5. Lookahead Adaptativo (1-ply nos 5 melhores candidatos; 2-ply quando em perigo)
              │
              ▼
 6. Avaliação da Peça no Hold (swap preventivo para S/Z/O sob terreno acidentado)
              │
              ▼
[Melhor Movimento Selecionado: Rotação Alvo, Coluna X Alvo e Uso de Hold]
```

1. **Geração de Candidatos**: Testa todas as 4 rotações de SRS e todas as translações horizontais ($X \in [-3, 9]$) onde a peça pode existir.
2. **Acessibilidade Física (`isReachable`)**: Simula se a peça consegue transitar lateralmente do ponto de spawn ($X=3$ ou $X=4$) até a coluna alvo sem ser bloqueada por espigões que atinjam o topo ($Y \le 0$).
3. **Simulação de Queda (`GetGhostY`)**: Desce a peça até o ponto mais baixo permitido no tabuleiro, grava o estado resultante em um tabuleiro clonado e calcula as linhas completadas.
4. **Avaliação Heurística (`evaluatePlacement`)**: Atribui uma pontuação escalar calculando relevo, buracos, altura, espigões e integridade do poço.
5. **Lookahead Adaptativo de 2 Camadas**:
   - **1-Ply**: Para os 5 melhores candidatos, simula a colocação da próxima peça conhecida (`NextQueue[0]`), combinando as notas ($Score_{\text{final}} = Score_{\text{cand}} + 0.65 \times Score_{\text{next}}$).
   - **2-Ply Preditivo**: Quando o tabuleiro entra em estado de atenção (altura $\ge 11$, irregularidade $\ge 14$ ou modo de limpeza ativo), uma segunda camada recursiva projeta também a terceira peça (`NextQueue[1]`). Isso permite à IA antecipar armadilhas e fendas que peças difíceis como `O`, `S` ou `Z` não conseguiriam preencher sem criar buracos.
6. **Decisão Inteligente de Hold**: Compara o melhor movimento com a peça atual contra o melhor movimento usando a peça no Hold (ou a primeira da fila caso o Hold esteja vazio).

---

### 2. A Estratégia 9-0 Stacking & As 4 Regras de Ouro do Poço

O empilhamento **9-0** constrói a estrutura do jogo exclusivamente nas colunas de **0 a 8**, preservando a **coluna 9** permanentemente vazia como um poço vertical dedicado para pontuar *Tetrises* de 4 linhas com a peça `I`.

Para impedir que a IA obstrua ou enterre o poço, quatro regras invioláveis foram codificadas na função de utilidade:

1. **Uso Exclusivo de Linha Vertical para Tetris**:
   - A peça `I` na orientação vertical na coluna 9 só é aceita se completar as **4 linhas completas** ($+180.000\text{ pts} + 90.000\text{ pts}$ de bônus).
   - Colocar uma peça `I` vertical na coluna 9 limpando menos de 4 linhas deixa blocos residuais formando uma parede na borda direita, recebendo uma penalidade catastrófica ($-120.000\text{ pts}$).
2. **Regra Anti-Espigão (Coluna 9 $\le$ Coluna 8)**:
   - A altura da coluna 9 nunca pode exceder a altura da coluna 8. Qualquer bloco excedente recebe penalidade proporcional severa ($-45.000\text{ pts} \times \Delta h$).
3. **Regra Anti-Teto / Anti-Fechamento**:
   - É terminantemente proibido colocar qualquer bloco na coluna 9 que deixe espaço vazio abaixo dele (teto sobre o poço). Esta condição recebe penalidade fatal ($-150.000\text{ pts}$), garantindo que peças futuras sempre alcancem o fundo.
4. **Controle de Profundidade e Acesso Lateral**:
   - A profundidade ideal do poço é mantida entre 2 e 4 blocos. Se a coluna 8 ficar mais de 4 blocos acima da coluna 9, a IA passa a penalizar aprofundamentos excessivos.
   - **Supressão do Espigão da Coluna 7**: A coluna 7 não pode formar um dente acima das colunas 6 e 8 ($-3.500\text{ pts} \times \Delta h$), assegurando que as peças possam deslizar suavemente até a coluna 8 e cair no poço sem impedimento.

---

### 3. Tratamento Especializado por Tipo de Peça (O, S, Z)

A análise minuciosa de mais de 2.000 partidas revelou que as peças `O`, `S` e `Z` eram responsáveis pela grande maioria dos *top-outs*. A IA agora incorpora regras dedicadas para cada uma:

- 🟨 **Peça O (2×2 - Plataforma Plana)**:
  - Como a peça `O` não rotaciona e possui largura 2, ela exige uma base plana de duas células adjacentes de mesma altura ($\Delta h = 0$).
  - **Bônus de Plataforma Nivelada**: Recebe $+3.500\text{ pts}$ quando encaixada sobre duas colunas perfeitamente niveladas.
  - **Penalidade de Degrau**: Penaliza severamente ($-3.500\text{ pts} \times \Delta h$) o encaixe sobre degraus de altura $\ge 2$, que deixariam a peça suspensa criando buracos ou espigões.
  - **Hold Preventivo para `O`**: Em terrenos com alta irregularidade ($\ge 12$) sem superfícies planas de 2 células, a peça `O` é automaticamente guardada no Hold em troca de uma peça flexível (`T`, `J`, `L`, `I`).
- 🟩 **Peça S e 🟥 Peça Z (Alinhamento Horizontal & Hold)**:
  - Peças diagonais causam problemas graves quando posicionadas verticalmente em espaços apertados, pois sua cauda inferior fica suspensa sobre vazios.
  - **Preferência Horizontal**: Posições horizontais recebem bônus ($+1.500\text{ pts}$), enquanto orientações verticais em terrenos irregulares são fortemente desencorajadas ($-3.500\text{ pts}$).
  - **Swap de Emergência no Hold**: Se uma peça `S` ou `Z` chega com o tabuleiro sob estresse (altura $\ge 9$, irregularidade $\ge 8$ ou presença de buracos), a IA troca preventivamente para o Hold para nivelar o terreno antes de posicioná-la.

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

$$\text{Limiar} = \text{clamp}\Big(14 - (3 \times \text{Buracos}) - \max(0, \text{AlturaCentral} - 8) - \max(0, \text{Bumpiness} - 7), \; 7, \; 15\Big)$$

- **Ativação Precoce**:
  - Se surgir o **primeiro buraco** na altura $\ge 8$, o modo de limpeza é ativado imediatamente para desenterrá-lo antes que ele seja coberto por novos blocos.
  - Se a irregularidade (*bumpiness*) atingir $\ge 13$ na altura $\ge 8$, o modo de limpeza atua para aplanar espigões antes que uma peça desfavorável cause *top-out*.
  - Se a altura central atingir $\ge 14$ ou a altura geral $\ge 16$, ativação de emergência máxima.
- **Diferenciação de Peça de Linha Imediata**:
  - Se a IA estiver com a peça `I` na mão e puder limpar 4 linhas no mesmo instante, o modo de limpeza é adiado por 1 turno para permitir a pontuação do Tetris (que reduz a pilha em 4 linhas instantaneamente).
  - Esperas por peças futuras da fila são proibidas se o centro estiver em risco ($\ge 11$ linhas).
- **Mudança Drástica de Pesos no Modo Limpeza**:
  - O custo de criar um buraco sobe de $15.000$ para **$55.000\text{ pts}$**, garantindo matematicamente que a IA nunca fure o tabuleiro para limpar uma única linha superficial.
  - Linhas parciais (1, 2 e 3 linhas) passam a receber grandes recompensas positivas ($+35.000$ a $+95.000\text{ pts}$) para baixar a altura rapidamente.

---

### 6. Mecânica de Lock Delay para Jogadores Humanos

Para partidas manuais humanas (`!AutoPlay`), implementamos a especificação oficial de **Extended Lock Down (Infinity Rule)**:
- Quando a peça toca o chão ou a superfície dos blocos, ela não trava imediatamente: um temporizador de tolerância de **500 ms** (2 ticks de gravidade) é concedido.
- Cada movimento horizontal ou rotação bem-sucedida reinicia o temporizador (com limite de segurança de até 15 reinícios).
- Pressionar **Espaço (Hard Drop)** trava a peça instantaneamente sem atraso.
- No modo Auto-Play, a IA desativa o lock delay para manter execução de máxima velocidade.

---

### 📊 Evolução Comparativa de Desempenho

Resultados empíricos consolidados comparando a versão original, a versão intermediária e a versão final otimizada:

| Métrica Avaliada | Baseline Original (1.000 jogos) | Intermediária (2.000 jogos) | Final Otimizada (519 jogos) | Evolução Geral |
| :--- | :---: | :---: | :---: | :---: |
| 💥 **Taxa de Tetris (4 Linhas)** | 41.4% | 53.3% | **60.0%** | **+18.6%** 🚀 |
| 💀 **Taxa de Derrota por Top-Out (Linhas 19-20)** | 33.8% | 33.2% | **14.8%** | **-56.2% de mortes** 🛡️ |
| 🏆 **Maior Pontuação Atingida** | 1.443.120 pts | 1.518.656 pts | **1.526.602 pts** | **+83.482 pts** |
| 🟨 **Mortalidade da Peça O** | 14.6% | 15.5% (#2) | **12.9%** | Suprimida |
| 🟩 **Mortalidade da Peça S** | 15.2% | 16.7% (#1) | **13.5%** | Suprimida |
| 🟥 **Mortalidade da Peça Z** | 15.9% (#1) | 14.8% | **13.5%** | Suprimida |
| ⚡ **Throughput de Simulação** | ~4.000 jogadas/s | ~7.000 jogadas/s | **> 9.200 jogadas/s** | **+130% de velocidade** |
| 📦 **Alocações no Heap por Jogada** | ~25 alocações | 1 alocação | **0 alocações na busca** | Zero-alloc no loop quente |

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
