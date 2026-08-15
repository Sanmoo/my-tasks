# 08 — Priorizar via $EDITOR

**What to build:** `mt prioritize` monta o buffer `[P]`/`[ ]` no formato da spec (linhas de Issue reordenáveis + comentários de instrução), inclui Issues `open` e `in_progress`, abre o `$EDITOR` e aplica: reordenar `[P]` muda a ordem; trocar `[ ]`↔`[P]` move entre a fila e o Backlog; ao salvar, os Ranks são renormalizados 1..N reescrevendo **apenas** os arquivos cujo rank mudou. Conteúdo inválido (ID inexistente ou duplicado) é rejeitado sem aplicar nada. O apply é **in-process** — nada de subprocesso por Issue (é a correção da lentidão original do nd-prioritize).

**Blocked by:** 03 — Criar, ver e editar Issues (schema completo)

**Status:** ready-for-agent

- [ ] Buffer no formato `[P]`/`[ ]` com comentários de instrução, incluindo `open` e `in_progress`
- [ ] Reordenar linhas `[P]` muda a ordem efetiva
- [ ] Toggle `[ ]`→`[P]` promove a Issue para a posição onde a linha foi colocada
- [ ] Toggle `[P]`→`[ ]` devolve a Issue ao Backlog
- [ ] Ranks renormalizados 1..N após salvar
- [ ] Apenas arquivos com rank alterado são reescritos (zero churn nos intactos)
- [ ] Buffer inválido → erro claro e nenhum arquivo alterado
- [ ] Apply in-process (sem subprocesso por Issue)
