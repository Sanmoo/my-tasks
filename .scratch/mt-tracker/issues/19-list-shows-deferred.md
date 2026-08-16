# 19 — list mostra todas as não-done (deferred inclusas)

**What to build:** `mt list` passa a exibir **todas** as Issues não-`done` — inclusive as com Deferred until no futuro, que hoje ficam ocultas por padrão. Inverte o item 18 do spec (`list` escondia deferred "para não ver trabalho que ainda não está disponível"). Decisões da sessão de grilling (17/…):

- Visibilidade: todas as deferred não-`done`, sem horizonte; só `done` fica oculta por padrão;
- Ordenação: posição natural de prioridade (rank → Backlog), misturadas — invariante compartilhado com `ready`/`overdue` preservado;
- Sufixo: `[defer MM-DD HH:MM]` (formato atual) passa a sair **sempre** nas deferred, no default e no `--all`; `--all` passa a significar apenas "inclui `done`";
- Escopo: só `list` muda; `ready`, `overdue`, `pick-next` e `--status`/`--label` intocados; sem flag nova de escape;
- Docs: spec item 18 reescrito, README, story/features e2e (`list.feature`, `defer.feature`), `CONTEXT.md` separando visibilidade de disponibilidade (sem ADR: default reversível).

**Blocked by:** 05 — list, 07 — Defer

**Status:** resolved

- [x] `internal/list`: `Visible()` deixa de ocultar future-deferred (só `done` oculta por padrão) + testes unit
- [x] `internal/cli`: sufixo `[defer ...]` fora do `if all`; help do `--all` e `listLong` atualizados
- [x] e2e: `list.feature` e `defer.feature` atualizados (story + cenários de ocultação invertidos); `ready-overdue.feature` verificado
- [x] Docs: spec item 18, README (seção `mt list`), `CONTEXT.md` (termo Deferral)
- [x] `make check` verde (unit + e2e + coverage ≥90% + mutation em `internal/list`) e finish no workflow de worktree

### Implementado

`Visible()` (internal/list) removeu a regra de ocultar future-deferred — só `done` oculta por padrão, e o parâmetro `now` (que ficou morto) saiu da assinatura junto. `runList` (internal/cli) moveu o sufixo `[defer MM-DD HH:MM]` para fora do `if all`: sai sempre nas deferred, no default e no `--all`; `--all` passou a significar só "inclui `done`". E2e: story e cenários invertidos em list.feature e defer.feature, e o cenário de sufixos combinados em dependency.feature agora cobre o default também; `ready-overdue.feature` verificado (ready continua excluindo deferred). Docs: spec item 18 invertido e item 27 ajustado ("tirar da fila acionável"), README reescrito, CONTEXT.md separa visibilidade de disponibilidade (sem ADR novo — default reversível). `make check` verde (unit + e2e + coverage ≥90%, list 100% + mutation). Revisado por /code-review (Standards sem violações; Spec conforme; achados tratados: parâmetro `now` morto removido, comment do `Options.All` corrigido; ticket 05 histórico deixado como registro, mesmo padrão dos ADRs).
