# 01 — resolução de vault pelo prefixo da issue ID

**What to build:** implementar o spec `key-prefix-resolution/spec.md` (Status: ready-for-agent): comandos com key resolvem o vault pelo ID prefix (`mt show dom-xyz` acha `@dom` sem `@`), precedência `@bookmark` > `--vault` > prefixo > default; not found no vault casado nomeia o vault; seleção explícita ganha hint; prefixo ambíguo erra listando bookmarks; `dep add` barra blocker de outro vault; fallback (sem `-` ou sem match) idêntico a hoje, mensagens inclusas.

**Blocked by:** —

**Status:** ready-for-agent

- [ ] `internal/vault`: `KeyVault`, `MatchByPrefix` (prefixo no primeiro `-`, case-sensitive, dedupe por path, config ilegível pulado) + testes unit black-box
- [ ] `internal/cli`: helper `resolveVaultForKey` + wiring em show, edit, done, reopen, status, defer, undefer (com id), top, bottom, rank, comment, dep add/rm
- [ ] Mensagens novas do spec (not found nomeando vault, hint, ambiguidade, same-vault no dep)
- [ ] e2e: `key-prefix.feature` com os cenários do spec
- [ ] Docs: `CONTEXT.md` (termo ID prefix)
- [ ] `make check` verde (coverage ≥90% + mutation em `internal/vault`) e finish no workflow de worktree
