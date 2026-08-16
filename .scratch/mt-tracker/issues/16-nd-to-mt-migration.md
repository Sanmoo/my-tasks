# 16 — Migração nd → mt (script + cutover)

**What to build:** Script Python idempotente `migrate-nd-to-mt.py` em `pkm-nd-migration/scripts/` (padrão dos scripts existentes: dry-run por padrão, `--apply` para escrever). Lê o vault nd vivo (`pkm/.vault/issues`) + sidecar `beads-unmigrated-sidecar.json` (fonte única dos `due_at`); escreve três vaults mt em `pkm/.vault/{bjd,dom,pes}` (cada um com `mt.yaml`: prefixo `bjd`/`dom`/`pes`; status `[open, in_progress, done]`, bjd adiciona `waiting`). Mapeamentos (ADR-0005): `closed` → `done` (+ `completed_at` = `closed_at`); `deferred` → `open` + `deferred_until` (Z/segundos → naive; só dia → `T00:00`); `waiting` → status customizado; `in_progress` preservada sem `started_at`; `due_at` do sidecar → `deadline`; comentários verbatim; IDs renomeados por domínio (`pkm-055` → `bjd-055`, case normalizado) com reescrita de referências no corpo via mapa global; campos descartados: `assignee`, `type`, `priority`, `content_hash`, `updated_at`, `created_by`, `close_reason`, `blocks`, `parent`; `blocked_by` preservado. Ordem: tudo no Backlog (labels de rank são histórico ambíguo). Relatório de auditoria em `migration-artifacts/` (mapeamento antigo→novo, status traduzido, deadlines, reescritas, descartes). Cutover (humano): último `nd sync`, congela `nd/backlog` (tag de backup), remove `.nd.yaml`, ajusta `.gitignore` do `.vault` para trackear issues no `main`, config global com `@bjd/@dom/@pes` + padrão, `mt check` nos três vaults, uma passada de `mt prioritize` por vault.

**Blocked by:** 15 — Dependências (schema `blocked_by` precisa existir antes do corte)

**Status:** ready-for-agent

- [x] Script lê vault nd vivo + sidecar (dry-run/apply, idempotente)
- [x] Tradução de status/datas conforme ADR-0005 (incl. `26-08-23` → `2026-08-23T00:00`)
- [x] `deadline` importado do sidecar (18 issues)
- [x] Comentários preservados byte a byte (marcadores beads mantidos)
- [x] IDs renomeados por domínio com reescrita de referências no corpo (mapa global, case normalizado)
- [x] Campos descartados removidos; `blocked_by` preservado
- [x] Relatório de auditoria gerado em `migration-artifacts/`
- [ ] Cutover executado: backup do `nd/backlog`, `.vault` reorganizado, config global, `mt check` OK, filas re-priorizadas

### Implementado (script, validado contra o vault vivo)

`pkm-nd-migration/scripts/migrate-nd-to-mt.py` (+ entry `migrate-nd-to-mt.py`, + 49 testes em `test_migrate_nd_to_mt.py`): dry-run por padrão, `--apply` para escrever, idempotente (2ª execução: 0 escritas, configs reusadas). Aplicado no vault vivo: 233 issues → `bjd` 177, `dom` 29, `pes` 27; 18 deadlines do sidecar; 176 `closed→done` (+`completed_at`), 14 `deferred→open` (+`deferred_until` naive, `26-08-23`→`T00:00`; `pkm-wm4` sem data), 5 `in_progress` sem `started_at`, 4 `waiting` (só bjd); 39 referências reescritas em 16 issues; 9 ids `PKM-*` normalizados; corpos byte a byte (marcadores beads, linhas `Due:` históricas); labels `rank/*`/`area/*` descartados (história ambígua, ADR), `kind/*` e `status/someday` mantidos. `mt check` (binário buildado do `main`): **OK nos três vaults**. Relatório: `pkm-nd-migration/migration-artifacts/migration-audit.json`.

### Code review (2 eixos, paralelos)

- **Standards:** sem violações duras (campo canônico do frontmatter, mt.yaml, padrão dos scripts locais, vocabulário do CONTEXT). Três avisos corrigidos: `closed_at` ausente em issue `closed` agora é `ValueError` com contexto (era KeyError cru); `load_sidecar` valida shape do JSON (id + lista `issues`); labels não-lista agora falha alto. Julgamentos mantidos: reescrita também dentro de URLs/spans de código (só ids mapeados, auditado — o mapa global é o mandato do ADR); `WORD_ID_RE` genérico em vez de prefixo fixo (deixa ids desconhecidos intactos).
- **Spec:** todos os itens verificados contra o vault vivo + sidecar. Nit corrigido: `defer_until` traduzido para `deferred_until` não é mais reportado como "descartado" no audit. Escopo extra deliberado (documentado): relatório de órfãos (nunca deleta), guarda de `blocked_by` entre vaults (falha alto; sem caso real), campos `related/led_to/follows/was_blocked_by` descartados (o mt os rejeitaria; regra geral do ADR).
- Verificação final: 49 testes OK, `ruff check` limpo, audit JSON válido, `mt check` OK em bjd/dom/pes, corpo de 233/233 issues idêntico ao fonte exceto reescritas.

### Cutover pendente (humano)

Ordem (per ticket + ADR-0005): último `nd sync` → tag de backup do `nd/backlog` → remover `.nd.yaml` e ajustar o `.gitignore` do `.vault` (issues voltam a ser trackeadas no `main`) → **reinstalar o `mt`** (`~/bin/mt` atual é build pré-ticket-15 e rejeita `blocked_by`) → config global com `@bjd/@dom/@pes` + padrão → `mt check` nos três vaults → uma passada de `mt prioritize` por vault (tudo migrou para o Backlog). Os três vaults já existem em `pkm/.vault/{bjd,dom,pes}` e passam no `mt check` hoje.
