# 12 — check e --fix

**What to build:** `mt check` audita o vault: Rank duplicado (erro, exit 1), lacunas de Rank (aviso brando), YAML/frontmatter malformado, Status fora da lista configurada do vault, formato de datetime inválido. `mt check --fix` renormaliza os Ranks para 1..N.

**Blocked by:** 06 — Transições de Status, 08 — Priorizar via $EDITOR

**Status:** ready-for-agent

- [ ] Rank duplicado reportado como erro (exit 1)
- [ ] Lacuna de Rank reportada como aviso brando
- [ ] YAML/frontmatter malformado detectado
- [ ] Status fora da config do vault detectado
- [ ] Datetime em formato inválido detectado
- [ ] `--fix` renormaliza Ranks 1..N
