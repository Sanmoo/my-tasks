# 10 — pick-next

**What to build:** `mt pick-next` escolhe entre Issues `open` e disponíveis (`now >= deferred_until`): menor Rank, senão a mais antiga do Backlog por `created_at` (desempate por ID); move para `in_progress` carimbando `started_at`. Sem candidatos: mensagem em stderr e exit 1. Rank duplicado no vault: recusa com erro claro. Múltiplas Issues `in_progress` simultâneas são permitidas (sem WIP limit).

**Blocked by:** 06 — Transições de Status, 07 — Deferir Issues, 08 — Priorizar via $EDITOR

**Status:** ready-for-agent

- [ ] Move a Issue correta conforme a ordem da spec
- [ ] Carimba `started_at` e seta `in_progress`
- [ ] Fallback para o Backlog mais antigo quando não há ranks
- [ ] Nada disponível → exit 1 + mensagem clara
- [ ] Rank duplicado → recusa com erro claro
- [ ] Várias Issues `in_progress` simultâneas funcionam
