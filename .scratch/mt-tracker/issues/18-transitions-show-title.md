# 18 — Transições mostram o título da Issue

**What to build:** a linha de confirmação das transições de status passa a exibir o título da Issue: `<id> is now <status>: <título>`. Vale para `pick-next`, `done`, `close`, `reopen` e `status` — todos caem no `applyMutation` compartilhado (internal/cli/status.go). O `dep` fica inalterado (`<id> is now blocked by <bloqueador>`, sem título). Decisões da sessão de grilling (17/…):

- Formato: sufixo com dois-pontos após o status — `pkm-002 is now in_progress: rank one`;
- Título vazio → sem sufixo (comportamento atual exato);
- Título multilinha → `\n` vira espaço, sem truncar;
- Asserções e2e reforçadas (pick-next.feature e status.feature) para incluir o título;
- README: exemplo do `pick-next` atualizado.

**Blocked by:** 10 — pick-next, 06 — Transições de Status

**Status:** ready-for-agent

- [ ] `applyMutation` imprime `<id> is now <status>: <título>` (com as regras acima)
- [ ] `pick-next.feature` e `status.feature` asseguram o título na saída
- [ ] README atualizado
