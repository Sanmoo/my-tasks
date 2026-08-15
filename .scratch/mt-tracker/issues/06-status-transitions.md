# 06 — Transições de Status

**What to build:** `mt done <id>` fecha a Issue (terminal, carimba `completed_at`), com `mt close` como alias; `mt reopen <id>` volta a `open` limpando `completed_at` e `started_at`; `mt status <id> <status>` transita livremente entre status validando contra a lista configurada do vault. Sem máquina de estados: apenas `pick-next`→`in_progress` e `done` terminal têm comportamento especial.

**Blocked by:** 03 — Criar, ver e editar Issues (schema completo)

**Status:** ready-for-agent

- [ ] `done` seta status `done` e carimba `completed_at`
- [ ] `close` comporta-se como alias de `done`
- [ ] `reopen` limpa `completed_at` e `started_at` e volta a `open`
- [ ] `status` aceita qualquer status da lista configurada do vault
- [ ] `status` rejeita valor fora da config com erro claro
- [ ] Nenhuma transição é barrada por máquina de estados
