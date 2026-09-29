# Histórico de alterações

## [1.0.0] — 29/09/2026 — Lançamento Oficial: Suporte Multilíngue (9 Idiomas), Empacotamento (.deb, .rpm, macOS) e Automação de Release

### Adicionado

- **Suporte Multilíngue Completo (i18n)** em `internal/i18n`:
  - 9 idiomas suportados nativamente:
    - 🇧🇷 Português do Brasil (`pt-br`, `pt`)
    - 🇺🇸 Inglês (`en`, `english`)
    - 🇪🇸 Espanhol (`es`, `spanish`)
    - 🇫🇷 Francês (`fr`, `french`)
    - 🇮🇹 Italiano (`it`, `ita`, `italian`)
    - 🇩🇪 Alemão (`de`, `german`, `deutsch`)
    - 🇷🇺 Russo (`ru`, `russian`)
    - 🇯🇵 Japonês (`ja`, `japanese`)
    - 🇨🇳 Chinês Simplificado (`zh`, `chinese`)
  - Troca dinâmica de idioma em tempo de execução via tecla `L` com menu modal interativo (teclas 1..9 ou setas).
  - Seleção de idioma via linha de comando através das flags `--lang=<idioma>` e `--idiom=<idioma>`.
  - Localização completa de cabeçalhos, estatísticas, painéis de controles, telas de aviso de terminal pequeno, menu de IA e overlays de Pausa e Game Over.
- **Empacotamento e Distribuição Multi-Plataforma**:
  - Pacotes Debian (`.deb`) para arquiteturas `amd64` e `arm64` (`dist/tetris_1.0.0_*.deb`).
  - Pacotes RPM (`.rpm`) para Fedora, RHEL e CentOS para arquiteturas `x86_64` e `aarch64` (`dist/tetris-1.0.0-1.*.rpm`).
  - Pacotes compactados `.tar.gz` para macOS (Apple Silicon `darwin_arm64` e Intel `darwin_amd64`) e Linux.
  - Pacote `.zip` para Windows (`windows_amd64`).
  - Geração de tabela de verificação criptográfica SHA-256 (`dist/checksums.txt`).
- **Automação de Build e Release**:
  - Script `scripts/build-release.sh` para compilação cruzada e geração automática de todos os pacotes usando `nfpm` e ferramentas nativas.
  - Script `scripts/release.sh` para criação de tag git `v1.0.0` e publicação de release via GitHub CLI (`gh release create`).
  - Script universal de instalação `scripts/install.sh` (`curl -fsSL ... | bash`) com auto-detecção de sistema operacional e arquitetura.
  - Fórmula Homebrew em `Formula/tetris.rb` para usuários de macOS e Linux.
  - `Makefile` com alvos `build`, `test`, `release`, `install` e `clean`.
  - Flags de versão `--version` e `-v` no executável informando a versão de lançamento e metadados de build.

## 28/09/2026 — Motor BitBoard, melhorias nas 3 políticas de IA e modelo de 50 atributos

### Adicionado

- Representação `BitBoard` (`[20]uint16`, 40 bytes na stack, zero ponteiros) em
  `internal/engine/board.go`, com tabelas de máscaras pré-computadas por peça,
  rotação e coluna (`piecePlacementMask`) e operações bitwise para colisão,
  queda fantasma, limpeza de linhas, alturas, buracos, bloqueios e transições.
- Cinco novos atributos estruturais e de lookahead no modelo de risco
  (`models/move-risk.json`, 50 atributos no total): `next_lookahead`,
  `hole_delta`, `top_hole_blockades`, `row_transitions` e `col_transitions`.
- Amostragem estratificada por nível de perigo em `./learn` (estados críticos,
  intermediários e normais) e contraste intra-decisão com candidatos topo,
  vice-líder e pior ranqueado.
- Proteção contra colapso de feixe em `selectBeam` (`internal/engine/beam.go`),
  preservando ao menos uma raiz (`depth = 0`) distinta quando `width >= 4`.

### Melhorado

- **Política v2 (`internal/engine/ai.go`)**:
  - Lookahead de 2 camadas (`NextQueue[0..1]`) sempre ativo desde a primeira
    jogada (viabilizado pelo custo sub-milissegundo do `BitBoard`) e expansão de
    `topLimit` de 5 para 7 candidatos.
  - Resgate via `Hold` quando a melhor colocação da peça atual criaria novos
    buracos e a peça do `Hold` evita a criação de buracos (`holdAvoidsHoleCreation`),
    além de guarda `!holdCreatesMoreHoles` em todas as trocas de `Hold`.
  - Paridade de alinhamento horizontal/vertical para `PieceZ` junto a `PieceS`,
    penalidade de delta de buracos (`holeDelta * 12000`), escavação rasa
    limitada (`shallowCover`) e proteção do corredor direito (`colunas 4..8 >= 14`)
    para passagem da peça `I` sob gravidade rápida (`80 ms`).
- **Política Híbrida (`internal/engine/learned.go`)**:
  - Rollouts contrafactuais em `SimulateRiskExamples` alinhados à execução
    `gameplay` (`StepAI` + `Tick` com gravidade real por nível).
  - Guarda de segurança em `findLearnedMove`: nunca sobrescreve para candidatos
    que criam mais buracos que o plano base, preserva a decisão de `Hold` da `v2`
    e atua apenas quando o plano base apresenta risco relevante.

### Corrigido

- Verificação de alcance horizontal no teto (`isReachableBitBoard`) restrita a
  `tryY ∈ {0, -1}`, removendo `tryY = -2` que permitia flutuar peças de 2 linhas
  inteiramente acima da linha `0`.

### Resultados e validação (Modo `gameplay`, limite de 2.000 jogadas, seeds `10001+`)

- **Latência e alocações (`go test -bench`)**:
  - Decisão `v2`: **1,77 ms → 0,44 ms** (**4,0× mais rápida**).
  - Decisão `beam (depth 10, width 4)`: **22,14 ms → 3,26 ms** (**6,8× mais rápida**,
    alocações reduzidas em **86%**, de `12,46 MB` para `1,76 MB/op`).
- **Política `v2` (150 partidas, 284.629 jogadas)**:
  - Sobrevivência até 2.000 jogadas subiu de **64,7% (`97/150`) → 87,3% (`131/150`)**
    (**-64% de derrotas**), `p10` de **630 → 1.753 jogadas**, taxa de Tetris de
    **58,01% → 70,70%** e média de buracos pós-drop de **0,14 → 0,03** (**-79%**).
- **Política `hybrid` (120 partidas, 229.225 jogadas)**:
  - Modelo retreinado com 6.640 rollouts de treino e 1.360 de validação em 135
    seeds (`Validation Brier: 0,07159` contra `0,15781` do prior, **-54,6%**).
  - Sobrevivência até 2.000 jogadas subiu de **65,8% (`79/120`) → 87,5% (`105/120`)**,
    superando a `v2` em média de jogadas/sessão (**1.910,2** vs **1.897,5**) e
    pior caso mínimo (**656** vs **86** jogadas), com **70,49%** de taxa de Tetris.
- **Busca em feixe `beam (depth 10, width 4)` (80 partidas, 160.000 jogadas)**:
  - Sobrevivência chegou a **100,0% (`80/80`, zero derrotas em 160.000 jogadas)**,
    taxa de Tetris subiu de **71,23% → 89,46%** e média de buracos caiu para **0,00**.

## 27/09/2026 — Modelo local de risco (experimental)

### Adicionado

- Modelo logístico portátil em `models/move-risk.json`, carregado uma vez na
  inicialização, com versão, atributos, limites de suporte e métricas de validação.
- Terceira opção no menu: **IA híbrida experimental**. Flags `-learned` e `-model`
  no jogo/simulador, com fallback para o planejador existente.
- Estimativa `P50 sim` no painel e probabilidades das alternativas nos logs. O
  alvo usa continuidade v2 por encaixe direto e futuros 7-bag amostrados.
- Snapshots com pose, fila, hold, gravidade, nível anterior à colocação e
  alternativas legais. O restante da bag é registrado sem sua ordem.
- Comando `./learn`: simulação de alternativas com futuros compartilhados,
  ajuste offline, validação por seed e publicação atômica do modelo.
- Artefato inicial com 45 atributos, treinado a partir de 120 decisões e 480
  combinações estado/ação: 1.552 rollouts de ajuste e 368 de validação.
  Treino e simulação executam na CPU; esta versão não inclui aceleração CUDA.
- Testes de serialização, entradas inválidas, suporte, fallback, seleção legal,
  reprodutibilidade e menu; relatório do piloto com novas seeds.

### Limites medidos

- Brier de validação **0,0215**, contra **0,0317** do prior constante, em 368
  rollouts de seis seeds separadas do ajuste.
- Gameplay em seeds 5001–5020, até 1.000 peças: **16/20** sobreviventes na híbrida,
  contra **19/20** na v2. Não houve desvios de aterrissagem.
- O primeiro modelo **não demonstrou melhora de gameplay**; a v2 continua padrão.
  Probabilidades são estimativas de um cenário simulado, não chances garantidas.

Detalhes: [modelo](models/README.md) e [validação](benchmarks/2026-09-27-risk.md).

## 26/09/2026 — Auto-Play reproduzível e lookahead experimental

### Adicionado

- Menu de IA no jogo: **M** abre a seleção entre a política v2 e o lookahead de
  **10 colocações**. ↑/↓ ou 1/2 escolhem; Enter confirma; Esc/M cancela.
  A partida fica congelada no menu. Alterar a política reinicia a rodada,
  preservando o recorde e o estado do Auto-Play.
- Busca em feixe experimental, com hold por camada, remoção de estados
  equivalentes e prévia ampliada para dez peças futuras. A profundidade inclui
  a peça atual; a prévia extra suporta hold vazio.
- Opções `-lookahead` e `-beam-width` no jogo e simulador. A política v2 continua
  como padrão; `-lookahead 10 -beam-width 4` habilita a alternativa experimental.
- Simulação reproduzível com `-seed`, modos `placement` e `gameplay`, resumo
  silencioso com `-quiet` e comparação de reserva rígida do poço com `-reserve-well`.
- Telemetria de planos, replanejamentos, profundidade e custo da busca, além de
  relatórios de benchmark e resumos JSON versionados.

### Melhorado

- Planejamento da peça atual com movimentos legais, rotações SRS nos dois
  sentidos, fase de gravidade e posições esperadas por ação.
- Replanejamento de rotas bloqueadas ou desatualizadas, substituindo as quedas
  cegas de watchdog; queda imediata quando a aterrissagem escolhida já é alcançável.
- Recuperação em emergência: penalidades do poço reduzidas para 8%, sem reduzir
  as penalidades por buracos e bloqueios.
- Documentação em português, com requisitos de compilação, controles, limites
  da IA, comandos de reprodução e resultados separados por experimento.

### Corrigido

- Contagem exata de partidas e resultados independentes da quantidade de workers.
- Classificação de top-outs, limites, cancelamentos, reinícios e sessões incompletas.
- Registro da peça que falhou, em vez de atribuir a derrota à última colocada.
- Recompensas incluindo soft/hard drop, sem pontos extras de linhas ocultas da busca.
- Avaliação do suporte da peça O usando o tabuleiro anterior à colocação.
- Medição do modo de limpeza na decisão e verificação da altura de aterrissagem.
- Propagação de falhas de gravação/flush e validação de JSON, sequência de jogadas,
  recompensas e totais no analisador.
- Finalização de workers e callback de progresso antes do retorno da simulação.

### Resultados e validação

- **1.000 seeds pareadas**, limite de 1.000 peças: sobrevivência de **41,1% para
  88,6%**, pontuação média **+49,4%**, zero desvios nas 951.555 colocações com plano.
  Participação de Tetrises caiu de 59,9% para 54,5%.
- **Piloto de 20 seeds pareadas**: lookahead com estados distintos chegou ao
  limite em **20/20**, contra 18/20 na v2; pontuação média **+11,37%** e participação
  de Tetrises de 55,6% para 70,7%. Zero desvios nas 20.000 colocações.
- O piloto não é uma nova validação com 1.000 partidas. A busca é aproximada,
  usa uma prévia maior e pode descartar a melhor sequência global. Médias de
  latência não garantem o pior caso; o simulador não mede atrasos reais da UI.
- Testes de regressão, detecção de corridas e análise estática:
  `go test ./...`, `go test -race ./...` e `go vet ./...`.

Consulte os [relatórios de benchmark](benchmarks/README.md) para metodologia,
comparações completas e comandos. Binários e datasets JSONL permanecem fora do Git;
recompile os executáveis depois de atualizar o código.
