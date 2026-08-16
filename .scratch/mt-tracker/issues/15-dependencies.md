# 15 — Dependências (blocked_by computado)

**What to build:** Campo `blocked_by: [<id>, ...]` no schema (mesmo vault, direção única: a issue registra quem a bloqueia). Uma issue está blocked se, e somente se, alguma issue listada não está `done` — estado computado, sem status e sem transição (padrão do ADR-0002, ver ADR-0004). Comandos: `mt dep add <id> <bloqueador>` e `mt dep rm <id> <bloqueador>`. Comportamento: `mt list` marca issues bloqueadas (sufixo `[blocked]`, no estilo do `[defer ...]`); `mt ready` e `mt pick-next` pulam issues blocked. `mt check` valida referências: ID existe no vault, sem auto-bloqueio, sem ciclos. Antecede a migração nd→mt (ticket 16).

**Blocked by:** 05 — Listar (padrão de sufixo/marcação), 07 — Defer (padrão de disponibilidade computada), 11 — Ready e overdue, 12 — Check (validações)

**Status:** ready-for-agent

- [x] Campo `blocked_by` no schema e no check de integridade
- [x] `mt dep add <id> <bloqueador>` / `mt dep rm <id> <bloqueador>` (erro de uso com argumentos malformados, exit 2)
- [x] Blocked computado: disponível ⇔ nenhum blocker em `blocked_by` não-done
- [x] `mt list` marca issues bloqueadas com `[blocked]`
- [x] `mt ready` e `mt pick-next` pulam issues blocked (sem nada disponível → exit 1 com mensagem)
- [x] `mt check` valida existência no vault, auto-bloqueio e ciclos
- [x] e2e cobrindo o fluxo completo (add → bloqueia → done do blocker → desbloqueia)

### Code review (2 eixos, paralelos)

- **Standards:** sem blockers. Dois ajustes aplicados: `scripts/audit-exit-codes.sh` ganhou probes do `dep` (o script documenta cobrir "every command"); a doc do package `list` citava as regras cobertas sem mencionar blocked. Julgamentos mantidos: predicado do query compartilhado carrega o mapa de status que só `ready` usa (o compartilhamento é o padrão existente); helper `blockedByItem` duplicado nos testes de list/check (padrão pré-existente de helpers por package).
- **Spec:** todos os itens verificados, com duas decisões registradas: (1) o parêntese "sem nada disponível → exit 1" vale para `pick-next` (story 44 da spec); `ready` mantém semântica de query — vazio = sucesso silencioso, como já coberto pelo e2e pré-existente de ready/overdue; (2) escopo extra deliberado: `dep add` valida existência/auto-bloqueio na hora (exit 1) e `dep rm` remove referências órfãs idempotentemente — ambos testados e documentados, a validação de ciclos permanece só no `check` como o ticket manda.
- Verificação final: `make check` passou (unit, e2e com 11 cenários novos de dependency, gate de cobertura ≥90% e mutação ≥90% — 2 mutantes novos do ciclo mortos por testes dedicados); `make audit` passou com os probes de dep.
