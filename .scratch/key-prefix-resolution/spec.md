Status: ready-for-agent

# Spec: resolução de vault pelo prefixo da issue ID

## Problem Statement

O usuário trabalha com vários Vaults, cada um com seu `prefix:` no `mt.yaml`, e as issue IDs carregam esse prefixo (`dom-xyz`, `pkm-055`). Mas ao referenciar uma Issue num comando (`mt show dom-xyz`), o vault é resolvido pela regra `@bookmark` > `--vault` > default bookmark — a ID em si não influencia nada. Se o default bookmark não for o `@dom`, o comando falha com "issue dom-xyz not found", e o usuário é obrigado a repetir o vault que a própria key já anuncia: `mt show @dom dom-xyz`. Redundância diária: a key identifica o vault por construção (o prefixo é o autor da verdade — `create` sempre gera IDs com o prefixo do vault), mas o CLI ignora essa informação.

## Solution

Nos comandos que recebem uma issue ID, o prefixo da key (a parte antes do primeiro `-`) passa a resolver o vault: `mt show dom-xyz` encontra o vault cujo `prefix:` é `dom`, sem precisar de `@dom`. A nova precedência é `@bookmark` > `--vault` > **prefixo da key** > default bookmark: seleção explícita continua vencendo exatamente como hoje, e o prefixo só substitui o degrau do default. Comandos sem key (`list`, `pick-next`, `check`, `create`, `q`, ...) ficam intocados.

## User Stories

1. As a usuário com o default bookmark apontando para `@pkm`, quero que `mt show dom-xyz` abra a Issue no vault `@dom`, para não precisar repetir o vault que a key já anuncia.
2. As a usuário, quero que `mt edit dom-xyz` abra o arquivo certo no vault `@dom`, para editar sem `@`.
3. As a usuário, quero que `mt done dom-xyz`, `mt reopen dom-xyz` e `mt status dom-xyz <status>` operem na Issue do vault `@dom`, para transições funcionarem sem `@`.
4. As a usuário, quero que `mt defer dom-xyz <when>` e `mt undefer dom-xyz` operem na Issue do vault `@dom`, para agendar sem `@`.
5. As a usuário, quero que `mt top dom-xyz`, `mt bottom dom-xyz` e `mt rank dom-xyz 2` operem na fila do vault `@dom`, para priorizar sem `@`.
6. As a usuário, quero que `mt comment dom-xyz <text>` adicione o comentário na Issue do vault `@dom`, para anotar sem `@`.
7. As a usuário, quero que `mt dep add pkm-123 dom-456` encontre o `<id>` no vault `@pkm` e o blocker no vault `@dom`, para referenciar Issues pelos seus prefixos.
8. As a usuário, quero que o prefixo seja a parte da key antes do primeiro `-` (ex.: `dom` em `dom-x-y`), para IDs com hífens extras continuarem funcionando.
9. As a usuário, quero que o match seja case-sensitive contra o literal do `prefix:` do `mt.yaml`, para o comando refletir exatamente a configuração do vault.
10. As a usuário, quero que uma key sem `-` (ex.: um arquivo herdado do vault default) continue resolvendo no vault default, para dados migrados seguirem acessíveis.
11. As a usuário, quero que `mt show @pkm dom-xyz` (seleção explícita) continue tentando no vault `@pkm` e falhe com not found quando a Issue não está lá, para um argumento explícito nunca ser desrespeitado por inferência.
12. As a usuário que errou de vault, quero que o not found do caso explícito ganhe um hint apontando o bookmark que o prefixo sugere (`hint: the prefix "dom" suggests vault @dom`), para eu ver o caminho certo sem um comando extra.
13. As a usuário, quero que o mesmo valha para `--vault <path>`: o flag explícito vence e o hint sugere bookmarks, para os dois mecanismos explícitos se comportarem igual.
14. As a usuário, quero que uma key cujo prefixo casa com um vault mas cujo arquivo não existe lá falhe nomeando o vault (`issue dom-zzz not found in @dom`), para eu saber que a resolução pulou para `@dom` e não procurou no default.
15. As a usuário com dois vaults usando o mesmo prefixo, quero um erro claro listando os bookmarks ambíguos (`key dom-xyz is ambiguous: bookmarks @a and @b both use prefix "dom" — pick one explicitly`), para nunca haver escolha silenciosa.
16. As a usuário, quero que dois bookmarks apontando para o mesmo vault **não** contem como ambiguidade, para aliases do mesmo diretório conviverem.
17. As a usuário, quero que um bookmark cujo `mt.yaml` não pode ser lido seja pulado na busca de prefixo (ele já é inutilizável para o comando), para um config quebrado não derrubar a resolução por prefixo.
18. As a usuário, quero que uma key cujo prefixo não casa com nenhum vault resolva no default bookmark com as mesmas mensagens de erro de hoje, para typos e arquivos herdados seguirem o comportamento atual.
19. As a usuário do `mt dep add`, quero que `<id>` e blocker em vaults diferentes falhem com erro claro (`blocker dom-456 belongs to @dom — blocked_by requires the same vault`), para a regra intra-vault do domínio continuar garantida mesmo com resolução por prefixo.
20. As a usuário do `mt dep rm`, quero que só o `<id>` resolva por prefixo (o blocker é apenas um nome a remover, como hoje), para a limpeza de referência obsoleta continuar idempotente.
21. As a usuário do `mt undefer` sem argumentos, quero que o lote continue operando no vault resolvido normalmente, para o hábito diário não mudar.
22. As a usuário, quero que todos os erros novos saiam com exit 1 (user error), para scripts distinguirem malformed (2) de estado (1) como sempre.
23. As a usuário, quero que o glossário ganhe o termo **ID prefix**, para o vocabulário do projeto nomear o conceito que agora dirige a resolução.

## Implementation Decisions

- **Termo novo no glossário (CONTEXT.md)**: **ID prefix** — a parte de uma issue ID antes do primeiro `-` (ex.: `dom` em `dom-xyz`); identifica o vault a que a Issue pertence: nos comandos que recebem uma key, quando nenhum `@bookmark` ou `--vault` é dado, o vault é resolvido pelo prefixo antes de recorrer ao bookmark default. `Avoid`: prefixo-do-vault (o termo já é esse).
- **Uma função pura exportada em `internal/vault`**: `MatchByPrefix(key, global, home)` → lista de candidatos `KeyVault{Bookmark, Path}`, zero quando a key não tem `-` ou nenhum prefixo casa, mais de um quando ambíguo. Extrai o prefixo no primeiro `-`, casa exatamente (case-sensitive) contra o `prefix:` de cada vault com bookmark, carregando os `mt.yaml` (pulando os ilegíveis) e dedupando por path expandido. Sem retorno de erro: tudo que pode falhar é pulado.
- **Precedência final**: `@bookmark` > `--vault` > prefixo da key > default. O prefixo consulta apenas bookmarks; o alvo de `--vault` nunca é candidato. O caminho de fallback (sem match) é byte a byte o comportamento de hoje, inclusive as mensagens de erro.
- **Wiring no `internal/cli`**: um helper compartilhado (`resolveVaultForKey`) usado por todos os comandos que recebem key — show, edit, done, reopen, status, defer, undefer (com id), top, bottom, rank, comment, dep add (id e blocker) e dep rm (só id). Comandos sem key continuam no `resolveVault` atual.
- **Mensagens** (erros saem com exit 1; o prefixo `Error: ` do runner é do formato padrão):
  - prefixo casou, arquivo ausente: `issue dom-zzz not found in @dom`
  - seleção explícita, arquivo ausente, prefixo casa com bookmark: `issue dom-xyz not found in @pkm` + `hint: the prefix "dom" suggests vault @dom`
  - prefixo ambíguo: `key dom-xyz is ambiguous: bookmarks @a and @b both use prefix "dom" — pick one explicitly: mt show @a dom-xyz`
  - dep cross-vault: `blocker dom-456 belongs to @dom — blocked_by requires the same vault`
- **`mt dep add`**: o vault do `<id>` é resolvido primeiro; o blocker, se o prefixo dele casar com um vault diferente, produz o erro de same-vault; senão é procurado no vault do `<id>` (regra atual).
- **Sem cache**: os `mt.yaml` dos bookmarks são lidos por invocação — escala de tracker pessoal, custo irrelevante.
- **Docs**: entrada nova no CONTEXT.md (texto da decisão acima); sem ADR — os três critérios (difícil de reverter, surpreendente sem contexto, trade-off real) não se sustentam: a decisão mora bem na spec e no glossário.

## Testing Decisions

- Os dois seams estabelecidos do repo (TESTING.md) cobrem tudo; **nenhum seam novo** e nenhuma mudança no Makefile — `internal/vault` já está matriculado nos gates (coverage ≥90% + mutation).
- **Seam 1 — processo CLI (primário)**: godog contra o binário compilado, feature nova `key-prefix.feature`. O harness já comporta múltiplos vaults: o segundo vault e o `config.yaml` (default + bookmarks) são escritos com o docstring step `the file "<path>" is written with:` (prior art: `bookmark.feature`). Cenários: caso feliz por família de comando (show/edit/status/dep); not found no vault casado por prefixo nomeando o vault; explícito `@` com hint; `--vault` com hint; prefixo ambíguo listando bookmarks; dois bookmarks para o mesmo path sem ambiguidade; key sem `-` e sem match no default com mensagens idênticas às de hoje; `dep add` com id e blocker de vaults diferentes; `dep rm` resolvendo só o id; `undefer` sem id no vault default. Prior art: `bookmark.feature`, `dependency.feature`, `issue.feature`.
- **Seam 2 — lógica pura exportada** (black-box, nunca internals): testes de unidade de `MatchByPrefix` em `internal/vault` com diretórios temporários (estilo dos testes de resolução existentes): match exato, case-sensitivity, primeiro `-` com hífens extras, sem `-`, sem match, ambiguidade, dedupe por path, config ilegível pulado. O wiring do cobra não tem gate de unidade — a e2e cobre, como todo o resto do wiring.

## Out of Scope

- Resolução por busca de arquivo (`issues/*.md` varrido por vault) — o prefixo é o autor da verdade.
- Busca por sufixo, fuzzy ou case-insensitive.
- Resolução por key em comandos sem key (`list`, `pick-next`, `check`, `create`, `q`, `overdue`, `ready`, `prioritize`, `bookmark`, `init`).
- Suporte a `blocked_by` entre vaults — a regra intra-vault permanece.
- Mudanças na gramática do `checkID` ou no formato de ID gerado por `create`.
- Cache de configs ou qualquer otimização de IO.
- Documentação no README — o glossário e a spec bastam por ora.

## Further Notes

- O design saiu de uma sessão de grilling (Q1–Q10, incluindo textos exatos de mensagens e a entrada do glossário), todas confirmadas pelo usuário; a entrega é um único ticket de implementação.
- Dados migrados seguem funcionando: uma ID no vault default cujo nome não segue o prefixo continua resolvendo lá (fallback).
- `mt show` hoje não lê o `mt.yaml`; com a feature, cada comando com key lê o config de todos os vaults com bookmark — por isso configs ilegíveis são pulados em silêncio, nunca erram.
- O custo mental do usuário cai: `mt <verbo> <key>` passa a ser suficiente em todos os vaults, e o `@` fica reservado para comandos sem key e para desambiguar.
