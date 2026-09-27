# Lookahead de dez colocações: implementação e viabilidade

A política experimental é opcional; `heuristic-9-0-v2` continua como padrão. Dez colocações **incluem a peça atual**, não dez peças adicionais. No jogo, **M** abre o menu; alterar a política reinicia a rodada e confirmar a já ativa não reinicia.

## Modelo de busca

- `-lookahead 10 -beam-width 4` habilita `beam-v2-depth10-width4` no jogo e no simulador.
- A fila cresce para dez peças futuras conhecidas. A peça extra suporta hold vazio. É uma variante com prévia maior, não previsão de um futuro aleatório invisível. Antecipar os sorteios preserva a sequência gerada; a UI mostra a fila como letras.
- Cada camada simula colocações com/sem hold, limpa linhas, rejeita colisões imediatas no nascimento da próxima peça e recalcula o modo de limpeza.
- As notas heurísticas são acumuladas com desconto de 0,85 por profundidade. Não são pontos previstos do jogo. O feixe conserva os estados distintos com melhor nota, usando ocupação, peça atual, hold e índice da fila como identidade. Rotações equivalentes não ocupam várias vagas. Empates estáveis tornam resultados independentes dos workers.
- Apenas a primeira colocação é executada. Sua rota considera gravidade; peças seguintes são replanejadas no tabuleiro real. Camadas futuras usam encaixes rápidos aproximados, não rotas completas de gravidade/SRS. Se nenhuma continuação alcançar o horizonte, utiliza-se o prefixo mais profundo disponível; sem candidato sobrevivente na primeira camada, usa-se a v2.
- `search_depth` informa a profundidade efetiva e `search_nodes` conta colocações avaliadas no plano registrado. Hold pode causar uma nova busca; os campos não medem todo o trabalho de CPU do turno.

Não é um otimizador exaustivo. Trinta opções em dez colocações já produzem `30^10 = 590.490.000.000.000` sequências antes dos ramos de hold. A poda pode descartar a melhor sequência global; ter a prévia não torna a busca estreita nem o modelo aproximado exatos.

## Experimento

Modo gameplay, seeds **101–120**, limite de **1.000 peças**, 20 partidas completas por política. A comparação v2 usa o subconjunto correspondente do benchmark anterior, não sua média global. A prévia não é amostrada e os pesos não foram ajustados nesse subconjunto. Pilotos pequenos não demonstram superioridade geral; sobrevivência significa atingir o limite, não vencer nem sobreviver indefinidamente.

### Busca final com estados distintos

| Métrica | v2, mesmas seeds | Feixe de dez colocações, largura 4 |
|---|---:|---:|
| Atingiram o limite | 18/20 (90%) | 20/20 (100%) |
| Pontuação média | 1.343.355,3 | 1.496.056,3 |
| Média de peças colocadas | 965,55 | 1.000 |
| Média de linhas | 381,35 | 396,6 |
| Participação de Tetrises nas linhas | 55,592% | 70,651% |
| Desvios de planos | 0 | 0 |

Pontuação média **+11,37%**. As seeds antes perdidas, 116 e 120, atingiram o limite; nenhuma regrediu em sobrevivência nessa amostra. Todas as **20.000** colocações corresponderam ao plano, e todos os planos registrados chegaram à profundidade dez. Zero replanejamentos e quedas de watchdog no agendamento nominal.

As buscas dos planos finais avaliaram, em média, 2.316,45 colocações (máximo 2.516), totalizando 46.329.031 avaliações. Buscas adicionais anteriores ao hold não entram nesses contadores.

JSONL: `logs/beam10-dedup-width4-20-seed101.jsonl`. Resumos versionados: [analisador](2026-09-26-lookahead-gameplay.json) e [comparação](2026-09-26-lookahead-comparison.json).

A execução levou 248,15 segundos com quatro workers e parte dos testes em paralelo; não é vazão isolada. **Não é um novo experimento com 1.000 partidas.** O subconjunto foi reutilizado no refinamento da deduplicação; ainda é necessária uma validação independente maior antes de mudar o padrão. A política também usa mais informação de prévia: a comparação não isola profundidade de disponibilidade de peças futuras.

JSONL ficam em `logs/`, fora do Git. A execução inicial interrompida de largura 8 foi excluída. O piloto sem deduplicação permanece apenas como comparação exploratória.

### Piloto inicial, sem deduplicação

| Métrica | v2, mesmas seeds | Feixe inicial, largura 4 |
|---|---:|---:|
| Atingiram o limite | 18/20 (90%) | 16/20 (80%) |
| Pontuação média | 1.343.355,3 | 1.281.036,05 |
| Média de peças colocadas | 965,55 | 935,8 |
| Média de linhas | 381,35 | 369,1 |
| Participação de Tetrises nas linhas | 55,592% | 58,087% |
| Desvios de planos | 0 | 0 |

O protótipo `beam-v1-depth10-width4` levou 219,69 segundos com quatro workers, verificações e outra execução concorrentes; não é vazão isolada. JSONL: `logs/beam10-width4-20-seed101.jsonl`.

## Latência no código final

Benchmarks Go em processo único, após terminar a simulação, em Intel Core i7-13620H, com um segundo por caso:

| Cenário / política | Tempo médio por decisão | Memória alocada por decisão |
|---|---:|---:|
| Tabuleiro vazio, v2 | 1,77 ms | 0,69 MB |
| Vazio, profundidade 10 / largura 4 | 22,14 ms | 12,46 MB |
| Vazio, profundidade 10 / largura 8 | 44,08 ms | 23,68 MB |
| Vazio, profundidade 10 / largura 16 | 82,14 ms | 46,45 MB |
| Nível 13, pilha irregular com buraco, v2 | 7,32 ms | 2,41 MB |
| Nível 13, pilha irregular com buraco, profundidade 10 / largura 4 | 20,81 ms | 11,43 MB |

Todos os casos de feixe chegaram a dez camadas. São médias em cenários fixos, não limites de pior caso nem percentis de latência. A largura 4 é viável nesses cenários frente ao intervalo de ação de 55 ms; largura 16 já o excede antes de uma possível busca adicional por hold. A busca vazia de largura 4 custa aproximadamente 12,5 vezes a v2, com pressão considerável de alocações.

## Validação

Passaram `go test ./...`, `go test -race ./...`, `go vet ./...` e `git diff --check`. Testes cobrem ranqueamento de uma camada por enumeração com hold, profundidade dez, execução, busca determinística sem mutação, deduplicação independente de cor sem unir holds/índices diferentes, prévias curtas, preservação da sequência, reinício, configuração inválida e gameplay reproduzível entre workers.

O menu tem cobertura de congelamento da partida, cancelamento sem mudança de política, confirmação sem reinício quando a opção já está ativa, preservação de pausa/recorde/Auto-Play e saída pelo teclado.

## Reprodução

Use arquivos novos: o logger acrescenta registros por padrão.

```bash
go run ./cmd/train -quiet -mode gameplay -lookahead 10 -beam-width 4 \
  -games 20 -workers 4 -seed 101 -max-moves 1000 -buf-kb 4096 \
  -file logs/beam-validation-fresh.jsonl
go run ./cmd/analyze_losses -file logs/beam-validation-fresh.jsonl -json
go test ./internal/engine -run '^$' -bench 'BenchmarkLookahead.*Decision' -benchtime=1s
go test ./...
go test -race ./...
go vet ./...
```

Timers virtuais **não** modelam atrasos reais de CPU. A busca interativa é síncrona e decisões caras podem atrasar renderização e tratamento da gravidade. Tempo no tabuleiro vazio não limita o pior caso; hold/replanejamento podem exigir mais de uma busca por callback.

Antes de adotar como padrão, testar amostras independentes maiores e latência interativa sob estresse. Ajustar avaliação e diversidade, não apenas profundidade. Tabuleiros compactos também podem reduzir alocações.
