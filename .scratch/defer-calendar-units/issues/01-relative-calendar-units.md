# 01 — Relativos em meses e anos no mt defer

**What to build:** `mt defer <id> <quando>` passa a aceitar `+<n>m` (meses)
e `+<n>y` (anos), além dos relativos exatos `+2d`/`+1w`/`+3h`. `m`/`y` usam
calendário: preservam o dia do mês alvo e o limitam ao último dia quando o
mês é mais curto (31/01 +1m → 28/02; 29/02 +1y → 28/02); a hora é mantida.
Um alvo além do ano 9999 — o maior que o `NaiveLayout` de quatro dígitos
representa — é rejeitado como duração inválida, sem overflow. Help do
comando, README e `defer.feature` listam as novas unidades.

**Blocked by:** None — can start immediately.

**Status:** resolved

- [x] `mt defer <id> +6y` grava `deferred_until` no mesmo dia/hora, N anos à frente
- [x] `mt defer <id> +6m` grava `deferred_until` em calendário, N meses à frente
- [x] Clamp de fim de mês: 31/01 +1m → 28/02 (29/02 em ano bissexto)
- [x] Clamp de ano bissexto: 29/02 +1y → 28/02
- [x] `+2d`/`+1w`/`+3h` continuam somando duração exata (sem regressão)
- [x] Alvo além de 9999 → erro de duração, sem overflow nem valor malformado
- [x] Help do `defer`, README e cenário e2e citam `m`/`y` — `make check` verde

## Comments

### Implementação

- `internal/deferral.Parse` aceita `m` (meses) e `y` (anos) no relativo
  `+<n><unit>`. `m`/`y` usam calendário (`addCalendar`): o dia do mês alvo é
  preservado e limitado ao último dia quando o mês é mais curto (31/01 +1m
  → 28/02; 29/02 +1y → 28/02); a hora é preservada. `d`/`w`/`h` continuam
  somando duração exata (`maxDuration`).
- Alvo além do ano 9999 (maior que o `NaiveLayout` de quatro dígitos
  round-trips) é rejeitado como `invalidDuration`; a checagem
  `years > maxYear-now.Year()` acontece antes da soma para que uma contagem
  perto do limite do `Atoi` não estoure o `int`.
- O catálogo de unidades das mensagens de erro virou a const `unitList`
  (uma fonte no pacote). Mensagens de args/help, README, `TESTING.md` e o
  header do `defer.feature` citam as novas unidades.
- Seam 2: unit table-driven com `+1y`/`+6y`, `+1m`/`+6m`/`+12m`/`+18m`,
  virada de ano, clamp de fim de mês e de 29/02, limites exatos
  (`+7973y`/`+7974y`, `+95680m`/`+95681m`) e Atoi-range para `m`/`y`.
  Seam 1: cenário no `defer.feature` adiando com `+6y` e `+6m`.
- Deltas de spec: nenhum funcional; a const `unitList` e a linha do
  `TESTING.md` são a dedup/atualização do catálogo.

### Code review (2 eixos, paralelos)

- Standards: nenhuma violação dura de padrão documentado; único achado
  acionável era o `TESTING.md` defasado (unidades `d/w/h`), corrigido. Os
  demais (catálogo repetido, shape de dispatch) eram judgement calls;
  dedupei o catálogo no pacote.
- Spec: compliant — nenhum requisito faltando, parcial ou comportamento
  não pedido; todos os limites foram re-traçados à mão.
- Verificação final: `make check` verde (unit, e2e, coverage 100% em
  `internal/deferral`, mutation gate); três mutantes equivalentes
  (`maxYear-now.Year()` redundante com o segundo guard e `day >= last`
  no-op) permanecem vivos, dentro do gate de 90%.
