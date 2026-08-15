# 15 — Dependências (blocked_by computado)

**What to build:** Campo `blocked_by: [<id>, ...]` no schema (mesmo vault, direção única: a issue registra quem a bloqueia). Uma issue está blocked se, e somente se, alguma issue listada não está `done` — estado computado, sem status e sem transição (padrão do ADR-0002, ver ADR-0004). Comandos: `mt dep add <id> <bloqueador>` e `mt dep rm <id> <bloqueador>`. Comportamento: `mt list` marca issues bloqueadas (sufixo `[blocked]`, no estilo do `[defer ...]`); `mt ready` e `mt pick-next` pulam issues blocked. `mt check` valida referências: ID existe no vault, sem auto-bloqueio, sem ciclos. Antecede a migração nd→mt (ticket 16).

**Blocked by:** 05 — Listar (padrão de sufixo/marcação), 07 — Defer (padrão de disponibilidade computada), 11 — Ready e overdue, 12 — Check (validações)

**Status:** ready-for-agent

- [ ] Campo `blocked_by` no schema e no check de integridade
- [ ] `mt dep add <id> <bloqueador>` / `mt dep rm <id> <bloqueador>` (erro de uso com argumentos malformados, exit 2)
- [ ] Blocked computado: disponível ⇔ nenhum blocker em `blocked_by` não-done
- [ ] `mt list` marca issues bloqueadas com `[blocked]`
- [ ] `mt ready` e `mt pick-next` pulam issues blocked (sem nada disponível → exit 1 com mensagem)
- [ ] `mt check` valida existência no vault, auto-bloqueio e ciclos
- [ ] e2e cobrindo o fluxo completo (add → bloqueia → done do blocker → desbloqueia)
