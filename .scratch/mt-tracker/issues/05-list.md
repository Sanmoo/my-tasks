# 05 — Listar Issues

**What to build:** `mt list` ordena por Rank (menor primeiro) → Backlog por `created_at` → ID; glyphs por Status (○ `open`, ◐ `in_progress`, ● `done`, fallback para custom); oculta Issues `done` e com Deferred until no futuro por padrão (deferidas mostram sufixo `[defer ...]` via `--all`); filtros `--status` e `--label`; warning quando há Rank duplicado no vault.

**Blocked by:** 03 — Criar, ver e editar Issues (schema completo)

**Status:** ready-for-agent

- [x] Ordem de listagem conforme a spec (rank → backlog por `created_at` → id)
- [x] Glyphs corretos por status, com fallback para status custom
- [x] `done` oculto por padrão e visível com `--all`
- [x] Issue com Deferred until futuro oculta (sufixo visível com `--all`)
- [x] Filtros `--status` e `--label` funcionam
- [x] Warning quando há Rank duplicado
- [x] Unit (Seam 2): comparador de ordenação coberto com alta cobertura

## Comments

### Implementação

- `internal/list` (Seam 2, gates: 100% coverage / 100% mutação): `Glyph`
  (○/◐/● + fallback `?` para custom), `Compare`/`Sort` (rank → backlog por
  `created_at` → id; `created_at` comparado lexicograficamente, que equivale
  a cronológico no stamp canônico zero-padded de `issue.NaiveLayout`),
  `IsFutureDeferred`/`DeferSuffix` (disponibilidade e sufixo `[defer MM-DD
  HH:MM]`), `Visible`+`Options` (regras de visibilidade: done/futura-deferida
  ocultas por padrão, `--all` revela, `--status`/`--label` filtram) e
  `DuplicateRanks`. `NaiveLayout` foi exportado de `internal/issue` (antes
  `createdAtFormat` duplicava o layout em `internal/cli`).
- `mt list`: lê `issues/*.md`, ordena, filtra via `Visible`, imprime
  `<glyph> <id>  <title>[ <suffix>]`; warning de rank duplicado vai ao
  stderr (calculado sobre o vault inteiro, antes de qualquer filtro).
- e2e `list.feature`: 8 cenários (ordem, glyphs+fallback, done oculto/`--all`,
  deferida-futura oculta + sufixo, filtros `--status`/`--label`, warning de
  rank duplicado, sem-vault). Reusa o passo `the file ... is written with:`
  do ticket 04 — nenhum passo novo.
- Deltas de spec: (1) o glyph de fallback não é nomeado na spec ("fallback
  para custom") — usei `?`. (2) `--status done` revela done sem `--all`
  (filtro de status explícito sobrepõe a ocultação padrão) — deliberado,
  assertado no e2e. (3) `--label` repetível é OR (any-match), não AND — a
  spec não define o combinador. (4) arquivo malformado faz `list` falhar
  com o ID nomeado (datetime, por outro lado, é leniente — `mt check` é o
  dono da validação de formato).

### Code review (2 eixos, paralelos)

- Spec: sem blockers; todos os checkboxes presentes e verificados. Três
  ambiguidades de baixa severidade (nenhuma exigiu mudança de código):
  `--status done` sobrepõe a ocultação padrão (deliberado, acima), `--label`
  OR, e `loadItems` falha forte em arquivo malformado enquanto datetime é
  leniente.
- Standards: sem violações hard. Corrigidos os achados: duplicação do layout
  de datetime (`NaiveLayout` exportado de `internal/issue`), invariante de
  `created_at` lexicográfico documentado no `Compare`, e a decisão de
  visibilidade movida do CLI para `internal/list.Visible` (Feature Envy →
  Seam 2 com cobertura unit, matando 22 mutantes).
