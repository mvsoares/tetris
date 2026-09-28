# Modelo local de risco — experimental

`move-risk.json` é um modelo logístico treinado offline, não uma tabela de
tabuleiros exatos nem frequências de escolhas da IA. A v2 continua como padrão.
O primeiro piloto da política híbrida regrediu em gameplay; o artefato é um
protótipo para estudo, não uma melhoria comprovada.

## Uso

Execute na pasta do projeto. `./tetris` tenta carregar `models/move-risk.json`
uma vez, e a opção **3** do menu **M** habilita a política híbrida. A ausência do
modelo deixa a opção indisponível. Com `-learned` explícito e modelo ausente ou
inválido, a CLI avisa e utiliza o fallback da política configurada.

```bash
./tetris -learned -model models/move-risk.json
```

Cada jogo compartilha o modelo imutável, inclusive entre workers. Reiniciar
preserva a seleção e o modelo. Editar o JSON durante uma partida não recarrega os
pesos: reinicie o programa. Caminhos relativos são resolvidos pela pasta atual.

## Significado da probabilidade

Alvo: `survive-v2-placement-sampled-7bag`, com horizonte de **50 colocações**, já
incluindo a ação candidata. A simulação respeita a fila conhecida, mistura apenas
a ordem não observada do restante da bag e usa novas bags 7-bag depois disso.
A continuidade executa a política v2 com simulação temporal de `gameplay`
(`StepAI` a cada `55 ms` + `Tick` de gravidade real do nível).

São 50 atributos: alturas, centro, buracos, irregularidade, ocupação, linhas da
candidata, nota heurística, limpeza, hold, nível, fase de gravidade, dez alturas
de coluna, tipos da peça candidata/próxima/hold e cinco atributos estruturais e
de lookahead (`next_lookahead`, `hole_delta`, `top_hole_blockades`,
`row_transitions`, `col_transitions`). As simulações usam a fila completa; os
atributos não codificam todas as peças futuras nem toda a geometria do tabuleiro.

Estimativas para várias alternativas **não somam necessariamente 100%**: cada
uma representa sobrevivência sob uma ação diferente. `P50 sim` não é chance
garantida de vitória. Intervalos de atributos são um filtro simples de suporte,
não prova de que um estado é conhecido ou que a previsão está bem calibrada.

## Treino e validação

`./learn` precisa de logs novos com snapshots `decision`. Logs antigos não
guardam todo o contexto; não são inventadas filas ou holds para completá-los.

- Apenas jogadas com plano correspondente de sessões completas são amostradas.
- São incluídos a opção registrada, o melhor encaixe imediato e alternativas
  amostradas. Ações não escolhidas **não** recebem automaticamente rótulo ruim.
- Futuros amostrados são compartilhados entre alternativas da mesma decisão.
  Seeds reais não são usadas para revelar a ordem invisível das peças.
- Seeds divisíveis por cinco ficam na validação; todas as alternativas e rollouts
  da mesma seed ficam juntos. Normalização usa só o treino.
- Ajuste logístico com rótulos binomiais, regularização L2 e gradiente em lote.
  O JSON inclui Brier, log loss e bins de calibração, contra um prior calculado no
  treino. Rollouts da mesma decisão são relacionados: não equivalem a estados
  ou partidas independentes.
- A publicação usa arquivo temporário e rename; falhas não substituem o modelo.

### CPU ou GPU?

O treino atual é implementado em Go e executa na CPU, assim como a geração dos
rollouts. O primeiro artefato tem apenas 45 pesos mais o intercepto, ajustados
com 388 exemplos agregados de treino. Não há integração CUDA: ter uma GPU NVIDIA
não acelera automaticamente esse código.

Para esse modelo pequeno, não há benefício demonstrado em migrar o ajuste para
GPU. A prioridade é alinhar os rótulos à execução gameplay e ampliar os estados
de treino. Uma futura rede neural com lotes maiores pode justificar treino em
GPU, mantendo a inferência do jogo portátil; isso ainda não está implementado.

## Política híbrida

A heurística calcula um plano legal primeiro. O modelo considera até seis
alternativas com melhores notas imediatas. Uma mudança exige previsão de pelo
menos 90% e vantagem superior a cinco pontos percentuais sobre o plano heurístico.
É uma regra de seleção, **não** garantia de confiança estatística ou segurança.

O fallback é mantido se o modelo for inválido, não superar o prior no Brier,
encontrar estado fora dos intervalos de treino ou não observar vantagem suficiente.
Superar o prior não certifica calibração nem melhora em gameplay. A UI permanece
síncrona e a híbrida ainda executa busca heurística; inferência barata não implica
uma decisão total mais rápida.

O identificador vem do JSON canônico, excluindo `id`. Formato, dimensões,
finitude, horizonte e separação de seeds são verificados na carga; arquivos acima
de 1 MB são rejeitados. Isso verifica compatibilidade, não autentica a origem.

## Primeiro artefato

Coleta: 30 partidas, seeds 1001–1030, até 1.000 peças, 29.165 decisões registradas.
Treino: 120 decisões, até quatro alternativas, quatro futuros por alternativa,
horizonte 50; 1.552 rollouts de ajuste e 368 de validação. Foram usadas 24 seeds
no ajuste e seis na validação. Jogar não altera o arquivo automaticamente.

Comandos e resultados estão no [relatório](../benchmarks/2026-09-27-risk.md).
