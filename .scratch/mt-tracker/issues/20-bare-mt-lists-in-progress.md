# 20 — mt sem comando lista as issues in_progress

**What to build:** invocar `mt` sem comando passa a executar a listagem de `mt list --status in_progress` contra o vault resolvido — o "o que estou fazendo agora", escondendo o backlog já priorizado. Hoje o bare `mt` mostra a ajuda raiz (comportamento ad hoc do esqueleto, não especificado). `mt help` e `mt --help` permanecem o caminho para a ajuda. Decisões da sessão de grilling:

- **Equivalência estrita**: a saída é byte a byte a de `mt list --status in_progress` — mesmas linhas (glyph, ID, título), mesmos sufixos `[blocked]`/`[defer ...]`, mesmo warning de Rank duplicado no stderr (calculado sobre o vault inteiro);
- **Gatilho**: nenhum comando restante após extrair `@bookmark` — cobre `mt`, `mt @bjd` e `mt --vault <path>`;
- **Vazio**: nenhuma issue `in_progress` → stdout vazio, exit 0, sem mensagem de dica;
- **Sem vault resolvível** (sem `@`, sem `--vault`, sem favorito default) → exit 1 com as instruções de resolução de vault; sem fallback para a ajuda;
- **Flags do `list` não sobem pro root**: `mt --status open` segue usage error exit 2;
- **Status personalizados nunca aparecem no bare `mt`** — a equivalência é com o status literal `in_progress`, mesmo que um vault defina custom status semanticamente similar;
- **Sem ADR** (default reversível; README + `rootLong` bastam como registro) e **sem mudança no CONTEXT.md** (vocabulário já cobre).

**Blocked by:** None — can start immediately.

**Status:** resolved

- [x] `mt` sem comando, com vault resolvível, imprime exatamente a saída de `mt list --status in_progress` (mesmas linhas e sufixos; mesmo warning de Rank duplicado no stderr)
- [x] `mt @bjd` e `mt --vault <path>` sem comando produzem a mesma listagem (gatilho = nenhum comando após extrair `@bookmark`)
- [x] Nenhuma issue `in_progress` → stdout vazio, exit 0, sem dica
- [x] Sem vault resolvível → exit 1 com instruções; sem fallback para a ajuda
- [x] `mt --status open` (e demais flags do `list` no root) segue usage error exit 2
- [x] `mt help` e `mt --help` inalterados: ajuda raiz, exit 0
- [x] Spec: item novo em "Implementation Decisions" definindo o comportamento bare; README atualizado (o trecho "`mt` sem argumentos mostra a ajuda raiz" sai); `rootLong` menciona o novo default
- [x] e2e: cenários novos — bare com vault listando só `in_progress`, bare com vault vazio (exit 0, stdout vazio), bare sem vault (exit 1), bare com `@bookmark`; smoke.feature intacto
- [x] `make check` verde (unit + e2e + coverage + mutation) e finish no workflow de worktree

### Implementado

O bare `mt` passou a executar `runList(cmd, false, "in_progress", nil)` — exatamente o código de `mt list --status in_progress`, garantindo equivalência byte a byte (linhas, sufixos `[blocked]`/`[defer ...]`, warning de Rank duplicado no stderr sobre o vault inteiro). `mt help` e `mt --help` seguem mostrando a ajuda raiz. Encontrado e corrigido um bug latente do harness: quando a linha de comando é só um `@bookmark` (ex.: `mt @bjd`), `BookmarkFromArgs` devolve `rest` nil e o cobra (SetArgs com nil = "não especificado") caía de volta em `os.Args[1:]` — reintroduzindo o `@bookmark` como comando desconhecido; `Run` agora normaliza nil → `[]string{}`. E2e: novo `bare.feature` com 8 cenários (só in_progress, equivalência com sufixos+warning, vault vazio com stdout vazio exit 0, sem vault exit 1, `@bjd`, default bookmark, flags do list não sobem pro root, `mt help`/`--help` inalterados) + steps `stdout` is empty e expansão de placeholders no conteúdo de `fileWrittenWith`. Docs: spec item novo em Implementation Decisions, README (seção `mt list` + `mt help`) e `rootLong` atualizados. Sem ADR e sem mudança no CONTEXT.md (default reversível). `make check` verde.
