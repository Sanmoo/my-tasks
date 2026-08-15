# 07 — Deferir Issues

**What to build:** `mt defer <id> <quando>` grava `deferred_until` aceitando `YY-MM-DD HH:MM` (com hora, sem truncar para data) e prazos relativos (`+2d`, `+1w`, `+3h`). A Issue permanece `open` mas indisponível enquanto `now < deferred_until`: some do `list` padrão e volta a aparecer sozinha quando o momento chega — sem operação de "undefer".

**Blocked by:** 05 — Listar Issues

**Status:** ready-for-agent

- [x] `defer` com `YY-MM-DD HH:MM` grava a hora (não truncada para data)
- [x] Relativos (`+2d`, `+1w`, `+3h`) calculados a partir de agora
- [x] Issue deferida permanece `open`
- [x] Issue deferida some do `list` e reaparece quando `now >= deferred_until`
- [x] Unit (Seam 2): parse de datetime (absoluto, relativos, bordas) com alta cobertura

## Comments

### Implementação

- `internal/deferral` (Seam 2, gates: 100% coverage / 100% mutação): `Parse`
  converte o argumento de tempo de `mt defer` no valor canônico
  `issue.NaiveLayout`. Absoluto `YY-MM-DD HH:MM` (a hora é preservada) e
  relativo `+<n><unit>` (`d` dias, `w` semanas, `h` horas) a partir de
  `now`; bordas (ano 00/68/69/99, mês 13, hora 25, minuto 60, `+0d`,
  unidade desconhecida, `++2d`, overflow) cobertas na unit.
- `issue.Defer(until)` grava só `DeferredUntil` — status e demais campos
  intocados (deferral é dado, não estado); unit asserta isso.
- `mt defer <id> <when>`: resolve o vault, faz o parse, escreve o campo e
  confirma com `<id> deferred until <canonical>`. O `when` junta os args
  restantes com espaço, então `26-08-20 08:00` (sem aspas) e
  `"26-08-20 08:00"` (com aspas) parseiam igual — mesmo padrão do título
  no `create`.
- e2e `defer.feature`: 5 cenários (absoluto preserva a hora, relativo,
  esconde/reaparece no `list`, tempo malformado → exit 1, sem tempo →
  usage exit 2). O esconder/reaparecer no `list` já existia (ticket 05 —
  `list.IsFutureDeferred`/`DeferSuffix`/`Visible`); o cenário amarra
  `defer` → `list`.
- Deltas de spec: (1) a expansão do ano de dois dígitos é `20YY` sempre —
  o `06` do `time.Parse` mapeia 69–99 para o século 1900 (sempre passado,
  inútil para defer); forço o século para que qualquer `YY` seja um alvo
  futuro deste século (`99` → 2099). (2) unidades relativas são
  exatamente `d`/`w`/`h` (as da spec); minutos/meses são rejeitados. (3)
  a mensagem de confirmação `<id> deferred until <canonical>` é decisão
  local (a spec não define o output do `defer`).
