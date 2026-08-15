# 05 — Listar Issues

**What to build:** `mt list` ordena por Rank (menor primeiro) → Backlog por `created_at` → ID; glyphs por Status (○ `open`, ◐ `in_progress`, ● `done`, fallback para custom); oculta Issues `done` e com Deferred until no futuro por padrão (deferidas mostram sufixo `[defer ...]` via `--all`); filtros `--status` e `--label`; warning quando há Rank duplicado no vault.

**Blocked by:** 03 — Criar, ver e editar Issues (schema completo)

**Status:** ready-for-agent

- [ ] Ordem de listagem conforme a spec (rank → backlog por `created_at` → id)
- [ ] Glyphs corretos por status, com fallback para status custom
- [ ] `done` oculto por padrão e visível com `--all`
- [ ] Issue com Deferred until futuro oculta (sufixo visível com `--all`)
- [ ] Filtros `--status` e `--label` funcionam
- [ ] Warning quando há Rank duplicado
- [ ] Unit (Seam 2): comparador de ordenação coberto com alta cobertura
