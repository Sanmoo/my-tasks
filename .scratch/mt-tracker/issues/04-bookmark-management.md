# 04 — Gerenciar bookmarks

**What to build:** `mt bookmark add <nome> <caminho>`, `mt bookmark list` e `mt bookmark rm <nome>` gerenciam os bookmarks e o favorito principal na config global auto-detectada. Um bookmark recém-adicionado resolve via `@nome` imediatamente; bookmark inexistente falha com mensagem clara.

**Blocked by:** 02 — Vault init, config e resolução de vault

**Status:** ready-for-agent

- [x] `add` grava o bookmark na config global
- [x] `list` mostra bookmarks e indica o favorito principal
- [x] `rm` remove sem corromper o restante da config
- [x] Bookmark recém-adicionado resolve via `@nome` no comando seguinte
- [x] `@nome` inexistente falha com mensagem clara e exit 1

## Comments

### Implementação

- `internal/vault` (Seam 2, gates: 97.1% coverage / 100% mutação): mutations
  puras `AddBookmark` (upsert, valida o nome), `RemoveBookmark` (limpa o
  `default` quando remove o favorito principal — nunca deixa um default
  pendurado), `IsValidBookmarkName` (gramática do `@nome` sem o `@`, a mesma
  de `bookmarkRe`), `Names` (ordem estável para `list`) e `SaveGlobal`
  (contraparte de escrita do `LoadGlobal`, cria diretórios pais, omite
  `default:`/`bookmarks:` vazios via `omitempty`). `TestAddThenResolveBookmark`
  cobre o round-trip add → save → load → resolve do item "resolve via `@nome`
  no comando seguinte".
- `mt bookmark add <nome> <caminho>` / `list` / `rm <nome>`: escrevem na config
  global auto-detectada (XDG). `list` imprime `nome -> caminho`, marcando o
  favorito principal com `(default)`. `rm` de bookmark inexistente falha com
  exit 1 ("bookmark @nome not found"); nome inválido no `add` é erro de uso
  (exit 2). Subcomandos recusam a forma `@nome` com mensagem clara ("take a
  bare name, not @nome") — o `@` é extraído globalmente antes do cobra, então
  um `@nome` posicional chegaria ao validador de aridade com a contagem errada.
- e2e: `bookmark.feature` com 8 cenários; passos novos `the file ... is written
  with:` (docstring) e `the file ... does not contain`; placeholders `<base>`/
  `<vault>` passam a expandir também nos argumentos das asserções
  `contains`/`does not contain`.
- Deltas de spec: (1) o item "resolve via `@nome` no comando seguinte" é
  provado em Seam 2 (round-trip); o e2e de processo da resolução (`@` >
  `--vault` > default) chega com o primeiro comando que exige vault (ticket
  03), mesmo delta registrado no ticket 02. (2) "configurar um favorito
  principal" (user story 11) não tem comando próprio — o `default:` é editado
  na config global (documentado no help e no TESTING.md); o ticket "What to
  build" lista só add/list/rm, então não foi criado `bookmark default`.

### Code review (2 eixos, paralelos)

- Spec: encontrado e corrigido — o cenário "add rejects an invalid bookmark
  name" usava `@pkm`, que é extraído globalmente e virava erro de aridade
  (passava por razão errada); agora usa `a/b` e asserta a mensagem, e um
  cenário novo cobre a recusa explícita da forma `@nome`. Confirmado que não há
  setter de default (delta documentado acima) e que o e2e de resolução fica
  para o ticket 03.
- Standards: `resolveVault` reutiliza `globalConfigPath` (remove a duplicação
  de detecção XDG); `AddBookmark`/`RemoveBookmark` compartilham
  `cloneBookmarks`; teste `TestBookmarkNameGrammarMatchesTokenMatcher` amarra as
  duas gramáticas de nome (a partilha via const foi descartada porque a
  concatenação de string derruba o gate de mutação do gremlins).
