# 18 — Transições mostram o título da Issue

**What to build:** a linha de confirmação das transições de status passa a exibir o título da Issue: `<id> is now <status>: <título>`. Vale para `pick-next`, `done`, `close`, `reopen` e `status` — todos caem no `applyMutation` compartilhado (internal/cli/status.go). O `dep` fica inalterado (`<id> is now blocked by <bloqueador>`, sem título). Decisões da sessão de grilling (17/…):

- Formato: sufixo com dois-pontos após o status — `pkm-002 is now in_progress: rank one`;
- Título vazio → sem sufixo (comportamento atual exato);
- Título multilinha → `\n` vira espaço, sem truncar;
- Asserções e2e reforçadas (pick-next.feature e status.feature) para incluir o título;
- README: exemplo do `pick-next` atualizado.

**Blocked by:** 10 — pick-next, 06 — Transições de Status

**Status:** resolved

- [x] `applyMutation` imprime `<id> is now <status>: <título>` (com as regras acima)
- [x] `pick-next.feature` e `status.feature` asseguram o título na saída
- [x] README atualizado

### Implementado

`applyMutation` (internal/cli/status.go) agora imprime via `transitionLine`: `<id> is now <status>` + `: <título>` quando a Issue tem título. Whitespace no título vira espaço único (`\r\n`/`\r`/`\n` → espaço, pontas podadas por `TrimSpace`) — linha de confirmação sempre única; título vazio/só-espaços mantém a saída anterior exata. O `dep` ficou inalterado (decisão Q5 do grilling). E2e: asserções com título em pick-next.feature (4), status.feature (2 + 2 cenários novos: título só-espaços e título multilinha) e dependency.feature (3); README com exemplo novo. `make check` verde (unit + e2e + coverage ≥90% + mutation 97.3%). Revisado por /code-review (Standards sem violações; Spec conforme; achados tratados: Replacer `\r`, vacua negativa removida, ticket resolvido).
