# 07 — Deferir Issues

**What to build:** `mt defer <id> <quando>` grava `deferred_until` aceitando `YY-MM-DD HH:MM` (com hora, sem truncar para data) e prazos relativos (`+2d`, `+1w`, `+3h`). A Issue permanece `open` mas indisponível enquanto `now < deferred_until`: some do `list` padrão e volta a aparecer sozinha quando o momento chega — sem operação de "undefer".

**Blocked by:** 05 — Listar Issues

**Status:** ready-for-agent

- [ ] `defer` com `YY-MM-DD HH:MM` grava a hora (não truncada para data)
- [ ] Relativos (`+2d`, `+1w`, `+3h`) calculados a partir de agora
- [ ] Issue deferida permanece `open`
- [ ] Issue deferida some do `list` e reaparece quando `now >= deferred_until`
- [ ] Unit (Seam 2): parse de datetime (absoluto, relativos, bordas) com alta cobertura
