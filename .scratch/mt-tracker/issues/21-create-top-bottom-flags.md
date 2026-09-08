# 21 — create/q com flags de entrada na fila

**What to build:** `mt create "<título>" -t/--top` e `-b/--bottom` criam a Issue já na fila de prioridade — no topo ou no fim — sem precisar rodar `top`/`bottom`/`prioritize` em seguida. `mt q` aceita as mesmas flags mantendo o contrato de imprimir só o ID. A semântica de ordem é a dos comandos de ordenação rápida (mesmo plano, uma só fonte de verdade); criar sem flag continua indo ao Backlog. As flags são mutuamente exclusivas.

**Blocked by:** None — can start immediately

**Status:** resolved

- [x] `create -b` põe a Issue no fim da fila reescrevendo apenas o arquivo novo (rank N+1; fila vazia → rank 1)
- [x] `create -t` põe a Issue na posição 1, deslocando a fila (resultado idêntico a criar sem flag e rodar `mt top <id>` em seguida)
- [x] `create -b` equivale a criar sem flag e rodar `mt bottom <id>`
- [x] `q -t`/`q -b` funcionam e imprimem apenas o ID
- [x] `-t` e `-b` juntas → erro de uso (exit 2)
- [x] `create` com flag imprime `Created <id> (rank <n>)`; sem flag, saída inalterada
- [x] Testes cobrem o caminho novo (unitários do pacote cli + e2e)

### Implementado

`mt create`/`mt q` ganharam `-t/--top` e `-b/--bottom` (mutuamente exclusivas — erro de uso, exit 2, nada é criado). A entrada na fila reusa o plano de ordenação rápida como única fonte de verdade: a nova Issue é planejada via `priority.QuickPlan` como se já existisse sem rank junto às Issues do vault (`planCreatePlacement`, internal/cli); os arquivos deslocados são reescritos antes do novo arquivo pousar (sem Rank duplicado transitório em disco) e a Issue nova nasce com o rank planejado — `-t` equivale a criar e rodar `mt top` (rank 1, fila deslocada), `-b` ao `mt bottom` (rank N+1; fila vazia → rank 1; com a fila contígua, só o arquivo novo é reescrito). Saída: `Created <id> (rank <n>)` com flag, `Created <id>` sem flag (inalterada), `q` imprime só o ID em todos os casos; criar sem flag segue para o Backlog. Testes: `internal/cli/issue_test.go` novo (caixa-preta via `cli.Run` — fim da fila, topo com deslocamento, equivalência create+top/bottom vs. flags, fila vazia, exclusão mútua com shorthand `-t/-b`, `q` silencioso) + 6 cenários e2e novos em `issue.feature` (a posição provada pela ordem do `mt list`). Docs: README (tabela, seção create/q, ajustes rápidos) e `Long` de create/q. Sem ADR e sem mudança no CONTEXT.md (vocabulário já cobre). `make check` verde. Revisado por /code-review (Spec conforme — 7/7 itens verificados; Standards sem violações — achados tratados: extração de `addPlacementFlags` no lugar das flags duplicadas, rank impresso direto do frontmatter, shorthand `-t/-b` coberto nos testes).
