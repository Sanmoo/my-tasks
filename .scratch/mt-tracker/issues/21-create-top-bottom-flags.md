# 21 — create/q com flags de entrada na fila

**What to build:** `mt create "<título>" -t/--top` e `-b/--bottom` criam a Issue já na fila de prioridade — no topo ou no fim — sem precisar rodar `top`/`bottom`/`prioritize` em seguida. `mt q` aceita as mesmas flags mantendo o contrato de imprimir só o ID. A semântica de ordem é a dos comandos de ordenação rápida (mesmo plano, uma só fonte de verdade); criar sem flag continua indo ao Backlog. As flags são mutuamente exclusivas.

**Blocked by:** None — can start immediately

**Status:** ready-for-agent

- [ ] `create -b` põe a Issue no fim da fila reescrevendo apenas o arquivo novo (rank N+1; fila vazia → rank 1)
- [ ] `create -t` põe a Issue na posição 1, deslocando a fila (resultado idêntico a criar sem flag e rodar `mt top <id>` em seguida)
- [ ] `create -b` equivale a criar sem flag e rodar `mt bottom <id>`
- [ ] `q -t`/`q -b` funcionam e imprimem apenas o ID
- [ ] `-t` e `-b` juntas → erro de uso (exit 2)
- [ ] `create` com flag imprime `Created <id> (rank <n>)`; sem flag, saída inalterada
- [ ] Testes cobrem o caminho novo (unitários do pacote cli + e2e)
