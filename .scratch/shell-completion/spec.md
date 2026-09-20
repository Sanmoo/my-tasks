Status: ready-for-agent

# Spec: shell completion de issue IDs

## Problem Statement

O `mt completion <shell>` da cobra já gera autocomplete de subcomandos e flags, mas nenhum comando define `ValidArgsFunction` — então as issue IDs (`dom-abcd`, `dom-abkg`) não completam. O usuário digita a key inteira à mão, mesmo quando só um prefixo já a determinaria: `mt show dom-abk<TAB>` deveria preencher `dom-abkg`, e `mt show dom-ab<TAB>` deveria listar `dom-abcd` e `dom-abkg`.

## Solution

Nos comandos que recebem uma issue key, um `ValidArgsFunction` compartilhado resolve o vault via `resolveVaultForKey` — a **mesma** função dos comandos, herdando `@bookmark` > `--vault` > prefixo > default e as regras de ambiguidade/fallback — e lista os IDs de `issues/*.md` filtrados por prefixo. Nenhuma lógica de resolução própria: completar e executar resolvem o vault de forma idêntica.

## User Stories

1. As a usuário, quero que `mt show dom-ab<TAB>` liste `dom-abcd` e `dom-abkg`, para ver as opções quando o prefixo é ambíguo.
2. As a usuário, quero que `mt show dom-abk<TAB>` complete `dom-abkg`, para não digitar a key inteira quando ela é única.
3. As a usuário, quero que `mt show dom-<TAB>` liste todas as issues do vault `@dom`, para explorar o vault pelo prefixo.
4. As a usuário, quero a mesma completação em todos os comandos que recebem key — `show`, `edit`, `done`, `reopen`, `status`, `defer`, `undefer`, `top`, `bottom`, `unrank`, `rank`, `comment`, `dep add`, `dep rm` — para o hábito valer em todo o CLI.
5. As a usuário do `mt dep add <id> <blocker>`, quero que o `<blocker>` complete com os IDs do **vault do sujeito** (resolvido pelo `<id>`), para a regra intra-vault do `blocked_by` continuar garantida; o mesmo vale para `dep rm`.
6. As a usuário com prefixo ambíguo (dois bookmarks com `dom`) ou sem vault resolvível, quero que o TAB devolva **zero candidatos** (nunca uma escolha silenciosa), para o erro claro do comando aparecer quando eu executar.
7. As a usuário, quero que a filtragem seja por prefixo case-insensitive do ID inteiro, para `DOM-AB<TAB>` e `dom-ab<TAB>` se comportarem igual sem afetar a resolução (que continua case-sensitive no `prefix:`).
8. As a usuário, quero que a instalação esteja documentada (`source <(mt completion zsh)` e equivalentes bash/fish), para ativar o completion sem surpresa.

## Implementation Decisions

- **Uma função de wiring compartilhada** em `internal/cli`: `completeIssueID(cmd, args, toComplete)` usada como `ValidArgsFunction` de todos os comandos com key. Não é lógica pura — é o mesmo tipo de wiring do restante do `internal/cli`, sem gate de unidade.
- **Vault**: `resolveVaultForKey(cmd, key)`. Na posição 0 do comando, `key = toComplete` (o prefixo parcial já basta: `PrefixOf("dom-ab")` = `dom`). Em `dep add`/`dep rm` na posição 1 (blocker), `key = args[0]` — o vault do sujeito, nunca o prefixo do blocker.
- **Candidatos**: `os.ReadDir(<vault>/issues)`, remove o sufixo `.md`, filtra por prefixo case-insensitive contra `toComplete`, retorna o ID inteiro com `ShellCompDirectiveNoFileComp` (sem completion de arquivos do shell).
- **Erro de resolução** (`resolveVaultForKey` devolve erro em ambiguidade ou sem vault) → devolve zero candidatos. O erro legível do comando aparece na execução, nunca durante o TAB.
- **Sem `-` no prefixo** (`dom<TAB>`) cai no default — comportamento herdado de `resolveVaultForKey`/`PrefixOf`. Não reimplementar uma extração "mais solta": reintroduziria o drift que o seam elimina.
- **Posições não-ID não completam**: o `<status>` de `mt status`, `<when>`, `<n>` de `rank`, `<text>` de `comment` e `@bookmark` ficam de fora (follow-ups). A posição 1 de `status`/`defer`/`rank`/`comment` devolve zero candidatos.
- **Docs**: README com o comando de instalação por shell (zsh primário, bash/fish documentados); o completion em si continua sendo o `mt completion <shell>` que a cobra já gera.
- **Sem termo novo no glossário**: completion é feature de implementação, não conceito de domínio. Nenhuma mudança em `CONTEXT.md` nem ADR.

## Testing Decisions

- **Seam 1 — processo CLI (primário)**: e2e godog contra o binário compilado, feature nova `completion.feature`, dirigindo `mt __completeNoDesc <cmd> <prefix>` (um candidato por linha — formato já verificado). Cenários: prefixo lista múltiplas; prefixo único completa uma; case-insensitive; `dom-` lista todas do vault; ambiguidade → stdout vazio; sem vault → stdout vazio; `dep add` blocker completa com o vault do sujeito; posição não-ID (`status` na posição 1) → vazio.
- **Sem seam novo, sem mudança no Makefile**: a listagem/filtro é wiring mecânica (mesma categoria do restante do cobra wiring, sem gate de unidade). O `resolveVaultForKey` já está coberto por `key-prefix.feature` e pelos unit de `internal/vault`.
- **Risco a verificar na implementação**: confirmar que a cobra parseia `--vault` durante o `__complete` (para o `resolveVaultForKey` lê-lo via `cmd.Flags()`). O `@bookmark` já está confirmado (o `__complete` passa por `Run`, que extrai e seta o `bookmark`). Se o flag não for parseado no completion, o fallback é lê-lo direto dos args do `__complete` — mudança local, sem afetar a decisão de reusar `resolveVaultForKey`.

## Out of Scope

- Completar o `<status>` de `mt status` (vault status list), `@bookmark`, `<when>`, `<n>`, `<text>`.
- Fuzzy, substring ou match por sufixo.
- Completar key sem `-` por prefixo solto (ex.: `dom<TAB>` → `dom-*`).
- Instalar/ativar o completion no shell do usuário automaticamente — só documentar.
- Alterar o formato de ID gerado ou o `checkID`.
- fish/powershell além da documentação do comando.

## Further Notes

- O design saiu de sessão de grilling (Q1–Q5 + A/B, todas confirmadas); a dependência `key-prefix-resolution` já está na `main`, e este ticket reusa `resolveVaultForKey`/`PrefixOf`/`MatchByPrefix` dela.
- O seam é a decisão central: completar e executar compartilham a mesma resolução de vault, então precedência, ambiguidade, hint e fallback nunca divergem.
- O `mt completion <shell>` já é gerado pela cobra (v1.10.2) sem código adicional; este trabalho só adiciona os `ValidArgsFunction`.
