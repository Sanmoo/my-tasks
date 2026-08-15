# 11 — ready e overdue

**What to build:** `mt ready` lista as Issues disponíveis agora (`open` + `now >= deferred_until`), em ordem de Rank; `mt overdue` lista Issues com Deadline estourado e não-`done`. Ambos com saída vazia (exit 0) quando nada se aplica.

**Blocked by:** 06 — Transições de Status, 07 — Deferir Issues

**Status:** ready-for-agent

- [ ] `ready` lista apenas disponíveis, ordenadas por Rank
- [ ] `overdue` lista Deadline estourado e não-`done`
- [ ] Saída vazia e exit 0 quando nada se aplica
- [ ] `overdue` respeita a natureza informativa do Deadline (não bloqueia nada)
