# 01 — shell completion de issue IDs

**What to build:** implementar o spec `shell-completion/spec.md` (Status: ready-for-agent): `ValidArgsFunction` nos comandos que recebem key, reusando `resolveVaultForKey` para resolver o vault e listando `issues/*.md` filtrados por prefixo (case-insensitive, ID inteiro, `ShellCompDirectiveNoFileComp`). `mt show dom-ab<TAB>` lista `dom-abcd`/`dom-abkg`; `dom-abk<TAB>` completa `dom-abkg`; `dom-<TAB>` lista todas do vault; ambiguidade/sem-vault devolvem zero candidatos; `dep add/rm` completam o blocker com o vault do sujeito; posições não-ID não completam; README documenta `source <(mt completion zsh)` (+ bash/fish).

**Blocked by:** —

**Status:** ready-for-agent

- [ ] `internal/cli`: helper `completeIssueID(cmd, args, toComplete)` — resolve o vault via `resolveVaultForKey` (posição 0 por `toComplete`; `dep` posição 1 por `args[0]`), lista `issues/*.md`, remove `.md`, filtra prefixo case-insensitive, retorna `ShellCompDirectiveNoFileComp`; erro de resolução → zero candidatos
- [ ] `ValidArgsFunction` nas 14 posições: show, edit, done, reopen, status, defer, undefer, top, bottom, unrank, rank, comment, dep add (id+blocker), dep rm (id+blocker)
- [ ] Posições não-ID (status na posição 1, `when`, `n`, `text`) sem candidatos
- [ ] e2e: `completion.feature` dirigindo `mt __completeNoDesc` com os cenários do spec (múltiplas, única, case-insensitive, `dom-`, ambiguidade, sem vault, dep blocker do sujeito, posição não-ID)
- [ ] Verificar `--vault` no `__complete` (cobra parseia flags no completion?); fallback: ler o flag dos args se preciso
- [ ] Docs: README — instalação do completion por shell (zsh primário, bash/fish)
- [ ] `make check` verde e finish no workflow de worktree
