# 04 — Gerenciar bookmarks

**What to build:** `mt bookmark add <nome> <caminho>`, `mt bookmark list` e `mt bookmark rm <nome>` gerenciam os bookmarks e o favorito principal na config global auto-detectada. Um bookmark recém-adicionado resolve via `@nome` imediatamente; bookmark inexistente falha com mensagem clara.

**Blocked by:** 02 — Vault init, config e resolução de vault

**Status:** ready-for-agent

- [ ] `add` grava o bookmark na config global
- [ ] `list` mostra bookmarks e indica o favorito principal
- [ ] `rm` remove sem corromper o restante da config
- [ ] Bookmark recém-adicionado resolve via `@nome` no comando seguinte
- [ ] `@nome` inexistente falha com mensagem clara e exit 1
