# 01 — overdue como atenção temporal + undefer

**What to build:** implementar o spec `overdue-undefer/spec.md` (Status: ready-for-agent): `mt overdue` vira o comando de atenção temporal — Deferrais expiradas primeiro (sufixo `[expirada MM-DD]`), depois Deadlines estourados (sufixo `[deadline MM-DD]`), cada grupo em ordem de Rank, Issue com os dois sinais só no grupo das expiradas, `done` fora, blocked dentro, vazio = exit 0; e `mt undefer` — sem args limpa todas as Deferrais expiradas do Vault (imprime `Undeferred <id> (was <valor>)` por Issue, zero expiradas = silêncio exit 0), com `<id>` limpa uma específica mesmo futura (sem campo = exit 1), tocando só o campo `deferred_until` (Status/Rank intocados). `ready`/`pick-next` e o `mt list --all` intocados.

**Blocked by:** —

**Status:** ready-for-agent

- [ ] `internal/list`: `DeferralExpired`, `ExpiredSuffix`, `DeadlineSuffix`, `OverdueGroups` + testes unit black-box
- [ ] `internal/issue`: `Issue.Undefer()` + teste
- [ ] `mt overdue` com runner próprio (dois grupos, sufixos, ordem de Rank)
- [ ] `mt undefer` batch + per-id com saídas e exit codes do spec
- [ ] e2e: `ready-overdue.feature` atualizado + feature nova de `undefer`
- [ ] Docs: `CONTEXT.md` (termo Deferral expirada), ADR-0006, README
- [ ] `make check` verde (coverage ≥90% + mutation em `internal/list` e `internal/issue`) e finish no workflow de worktree
