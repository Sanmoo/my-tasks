# 19 — list mostra todas as não-done (deferred inclusas)

**What to build:** `mt list` passa a exibir **todas** as Issues não-`done` — inclusive as com Deferred until no futuro, que hoje ficam ocultas por padrão. Inverte o item 18 do spec (`list` escondia deferred "para não ver trabalho que ainda não está disponível"). Decisões da sessão de grilling (17/…):

- Visibilidade: todas as deferred não-`done`, sem horizonte; só `done` fica oculta por padrão;
- Ordenação: posição natural de prioridade (rank → Backlog), misturadas — invariante compartilhado com `ready`/`overdue` preservado;
- Sufixo: `[defer MM-DD HH:MM]` (formato atual) passa a sair **sempre** nas deferred, no default e no `--all`; `--all` passa a significar apenas "inclui `done`";
- Escopo: só `list` muda; `ready`, `overdue`, `pick-next` e `--status`/`--label` intocados; sem flag nova de escape;
- Docs: spec item 18 reescrito, README, story/features e2e (`list.feature`, `defer.feature`), `CONTEXT.md` separando visibilidade de disponibilidade (sem ADR: default reversível).

**Blocked by:** 05 — list, 07 — Defer

**Status:** ready-for-agent

- [ ] `internal/list`: `Visible()` deixa de ocultar future-deferred (só `done` oculta por padrão) + testes unit
- [ ] `internal/cli`: sufixo `[defer ...]` fora do `if all`; help do `--all` e `listLong` atualizados
- [ ] e2e: `list.feature` e `defer.feature` atualizados (story + cenários de ocultação invertidos); `ready-overdue.feature` verificado
- [ ] Docs: spec item 18, README (seção `mt list`), `CONTEXT.md` (termo Deferral)
- [ ] `make check` verde (unit + e2e + coverage ≥90% + mutation em `internal/list`) e finish no workflow de worktree
