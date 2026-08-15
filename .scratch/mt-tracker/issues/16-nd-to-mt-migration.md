# 16 — Migração nd → mt (script + cutover)

**What to build:** Script Python idempotente `migrate-nd-to-mt.py` em `pkm-nd-migration/scripts/` (padrão dos scripts existentes: dry-run por padrão, `--apply` para escrever). Lê o vault nd vivo (`pkm/.vault/issues`) + sidecar `beads-unmigrated-sidecar.json` (fonte única dos `due_at`); escreve três vaults mt em `pkm/.vault/{bjd,dom,pes}` (cada um com `mt.yaml`: prefixo `bjd`/`dom`/`pes`; status `[open, in_progress, done]`, bjd adiciona `waiting`). Mapeamentos (ADR-0005): `closed` → `done` (+ `completed_at` = `closed_at`); `deferred` → `open` + `deferred_until` (Z/segundos → naive; só dia → `T00:00`); `waiting` → status customizado; `in_progress` preservada sem `started_at`; `due_at` do sidecar → `deadline`; comentários verbatim; IDs renomeados por domínio (`pkm-055` → `bjd-055`, case normalizado) com reescrita de referências no corpo via mapa global; campos descartados: `assignee`, `type`, `priority`, `content_hash`, `updated_at`, `created_by`, `close_reason`, `blocks`, `parent`; `blocked_by` preservado. Ordem: tudo no Backlog (labels de rank são histórico ambíguo). Relatório de auditoria em `migration-artifacts/` (mapeamento antigo→novo, status traduzido, deadlines, reescritas, descartes). Cutover (humano): último `nd sync`, congela `nd/backlog` (tag de backup), remove `.nd.yaml`, ajusta `.gitignore` do `.vault` para trackear issues no `main`, config global com `@bjd/@dom/@pes` + padrão, `mt check` nos três vaults, uma passada de `mt prioritize` por vault.

**Blocked by:** 15 — Dependências (schema `blocked_by` precisa existir antes do corte)

**Status:** ready-for-agent

- [ ] Script lê vault nd vivo + sidecar (dry-run/apply, idempotente)
- [ ] Tradução de status/datas conforme ADR-0005 (incl. `26-08-23` → `2026-08-23T00:00`)
- [ ] `deadline` importado do sidecar (18 issues)
- [ ] Comentários preservados byte a byte (marcadores beads mantidos)
- [ ] IDs renomeados por domínio com reescrita de referências no corpo (mapa global, case normalizado)
- [ ] Campos descartados removidos; `blocked_by` preservado
- [ ] Relatório de auditoria gerado em `migration-artifacts/`
- [ ] Cutover executado: backup do `nd/backlog`, `.vault` reorganizado, config global, `mt check` OK, filas re-priorizadas
