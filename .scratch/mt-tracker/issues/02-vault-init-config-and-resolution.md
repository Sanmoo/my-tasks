# 02 — Vault init, config e resolução de vault

**What to build:** `mt init [dir]` cria um Vault utilizável (diretório de issues + config do vault com prefixo de ID e lista de Status); a config global é detectada automaticamente (XDG) com bookmarks e favorito principal; resolução de vault na precedência `@bookmark` > `--vault <path>` > favorito principal; sem nenhum deles, o comando falha em stderr com instruções (exit 1). Status default (`open`, `in_progress`, `done`) valem quando a config não os define.

**Blocked by:** 01 — Esqueleto do projeto e seams de teste

**Status:** ready-for-agent

- [x] `mt init` cria vault utilizável com config default
- [x] init aceita prefixo de ID e lista de status customizada
- [x] Config global com bookmarks e favorito principal é lida automaticamente
- [x] Resolução respeita a precedência `@` > `--vault` > `default`
- [x] Sem nenhum definido, qualquer comando falha com mensagem de instrução e exit 1
- [x] Unit (Seam 2): precedência de resolução coberta com alta cobertura

## Comments

### Implementação

- `internal/vault` (Seam 2, gates: 97.2% coverage / 100% mutação): config global
  (`LoadGlobal`, `GlobalConfigPath` XDG), config do vault (`LoadVault`, `Save` —
  prefixo + status com defaults `[open, in_progress, done]`, `status,flow` no
  YAML), resolução `Resolve` (@ > `--vault` > default; erro com instruções
  quando nada definido), `BookmarkFromArgs` (token `@nome` em qualquer posição
  dos args, no máximo um), `ExpandHome`, `PrefixFor` (prefixo derivado do nome
  do diretório, ≤8 chars, fallback `--prefix` quando não derivável).
- `mt init [dir]`: cria `issues/` + `mt.yaml`; `--prefix` (default derivado do
  dir) e `--status` repetível (default: open/in_progress/done); recusa vault
  existente (exit 1); não aceita @bookmark (exit 2). `Vault ready at <path>`
  no stdout.
- CLI: `--vault` como flag persistente; extração do @ antes do cobra via
  `SetArgs` (corrige o `Run` do esqueleto, que executava `os.Args[1:]` mesmo
  com args diferentes); helper `resolveVault` compartilhado para os comandos
  que exigem vault (tickets 03+).
- e2e: `init.feature` com 5 cenários; steps novos (file/dir exists, file
  contains) e placeholders `<base>`/`<vault>` no step `I run`.
- Makefile: `gremlins unleash` aceita 1 package por invocação — `mutate`
  expande um comando por package (`$(foreach)`), com `.SILENT`.
- Deltas de spec: (1) a resolução de vault é coberta em Seam 2 (unit), como o
  próprio ticket pede no último item; o e2e de processo da resolução (falha
  sem vault, precedência) chega com o primeiro comando que exige vault
  (ticket 03). (2) `@nome` pode aparecer em qualquer posição dos args
  posicionais; nomes de bookmark não podem começar com `@` — regra
  documentada no TESTING.md.

### Code review (2 eixos, paralelos)

- Spec: blocker encontrado e corrigido — `mt init` sem dir nunca funcionava
  (`PrefixFor(".")` = ""). Agora resolve o cwd via `os.Getwd()` e deriva o
  prefixo do nome do diretório corrente; cenário e2e novo cobre o caminho
  default-dir. `init` também passou a recusar `--vault` (exit 2), consistente
  com a recusa de @bookmark.
- Standards: doc do package alinhado (`default favorite` → `default bookmark`)
  e `TrimLeft` redundante removido do `PrefixFor`. `resolveVault` permanece
  sem chamador até o ticket 03 (documentado); harness ganhou `RunCmdIn` +
  step de working directory para o cenário de bare init.
