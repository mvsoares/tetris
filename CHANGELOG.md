# Histórico de alterações

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
