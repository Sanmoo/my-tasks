# 09 — Ajustes rápidos de ordem

**What to build:** `mt top <id>`, `mt bottom <id>`, `mt rank <id> <n>` e `mt unrank <id>` ajustam a ordem sem abrir editor, reusando a renormalização 1..N do prioritize. `unrank` devolve a Issue ao Backlog.

**Blocked by:** 08 — Priorizar via $EDITOR

**Status:** ready-for-agent

- [ ] `top` move a Issue para o primeiro lugar da fila
- [ ] `bottom` move para o fim da fila
- [ ] `rank <n>` insere na posição n, deslocando e renormalizando
- [ ] `unrank` remove o rank (Issue vai para o Backlog)
- [ ] Comandos reescrevem apenas os arquivos cujo rank mudou
