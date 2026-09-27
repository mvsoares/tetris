# Benchmarks do Auto-Play

Os relatórios estão em português; nomes de políticas, opções de CLI e campos JSON
mantêm os identificadores usados no código.

| Relatório | Comparação | Amostra |
|---|---|---:|
| [Diagnóstico inicial](2026-09-26.md) | Encaixe direto versus movimentos e gravidade, política v1 | 1.000 seeds por modo |
| [Implementação e reteste da v2](2026-09-26-v2.md) | v1 versus v2, mesmas seeds e tempos nominais | 1.000 seeds por política |
| [Viabilidade do lookahead de 10 colocações](2026-09-26-lookahead.md) | v2 versus busca em feixe; piloto e latência | 20 seeds pareadas |
| [Modelo local de risco](2026-09-27-risk.md) | Validação por seed de probabilidades e piloto v2 versus híbrida | 368 rollouts de validação; 20 novas seeds |

Todas as comparações de sobrevivência usam um limite de 1.000 peças. Atingir o
limite não significa vencer nem sobreviver indefinidamente. `gameplay` usa timers
virtuais: atrasos reais de CPU, renderização e mensagens na UI não são simulados.

Os JSON de resumo ficam versionados junto aos relatórios. Os datasets completos
ficam em `logs/` e são ignorados pelo Git. Use um caminho novo em cada reprodução:
o logger acrescenta registros ao arquivo existente por padrão. Com o código atual,
as execuções padrão usam a v2; os comandos históricos da v1 exigem a implementação
antiga e não recriam aquela política no código atual.
