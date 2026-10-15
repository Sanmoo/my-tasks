# Issue tracker: `mt` (vault `myt`)

Issues e specs deste repo vivem no vault `myt` do próprio repo — o diretório
`issues/` na raiz, com um arquivo Markdown por Issue e frontmatter YAML, operado
pelo CLI `mt` (o tracker é o produto deste repo; dogfooding). Especificações são
publicadas como o corpo (Description) de uma issue com a label `ready-for-agent`.

O vault é resolvido pelo **prefixo do ID** (`myt-…`) ou por `@bookmark` — **nunca
pelo cwd**: sem `--vault` explícito, o bookmark *default* vence e a issue cai no
vault errado (isso já aconteceu). Prefira a resolução pelo prefixo do ID ou um
bookmark dedicado.

> Tickets criados antes da migração ficam em `.scratch/` (nos arquivos) e issues
> do GitHub ficam como histórico; nada é movido. Issues novas vão para o vault
> `myt`.

## Convenções

- **Criar uma issue**: `mt create "<título>" --label <labels...>` (ou `mt q` quando
  só o ID interessa). `create` roda a sessão de encaixe quando há terminal; em pipe
  a issue fica no Backlog. O ID (`myt-xxxx`) é gerado na criação — como edges de bloqueio precisam do ID,
  elas são gravadas num segundo passo (`mt dep add`).
- **Ler uma issue**: `mt show <id>` (completo) ou `mt show <id> -s` (resumo).
  Comments fazem parte do corpo e aparecem no `show`.
- **Listar**: `mt list` (ordem de Fila) com `--status` e `--label` (repetível)
  para filtrar. O bucket "sem label" do triage é filtrado client-side a partir
  de `mt list --all`.
- **Comentar**: `mt comment <id> "<texto>"` (append-only, âncora estável).
  Transições aceitam comentário à direita (`mt done <id> "<texto>"`).
- **Aplicar/retirar labels**: não há comando — edite o frontmatter diretamente
  (`labels: [...]`) e valide com `mt check`. Na criação, use `--label` (repetível).
- **Fechar**: `mt done <id> "[comment]"`. **Reabrir**: `mt reopen <id>`.
- **setStatus livre**: `mt status <id> <status> "[comment]"` — qualquer status na
  lista do vault (`open`, `in_progress`, `done`).

Inferir o vault pelo prefixo do ID da key (`myt-…` → este repo).

## Pull requests as a triage surface

**PRs as a request surface: no.** O `mt` não é um tracker de PRs — PRs continuam
no GitHub (`gh pr …`), fora do escopo do triage.

## Quando uma skill diz "publish to the issue tracker"

Crie uma issue no vault `myt`: `mt create "<título>" --label <labels...>` e
escreva o conteúdo (spec, brief, resposta) na seção Description do arquivo, com
`mt check --vault <vault>` para validar.

## Quando uma skill diz "fetch the relevant ticket"

`mt show <id>` — corpo completo (Description, Notes, Comments) e frontmatter.

## Wayfinding operations

Usado por `/wayfinder`. O **map** é uma issue com **child** issues como tickets.

- **Map**: uma única issue com a label `wayfinder:map`, corpo com Destination /
  Notes / Decisions-so-far / "Not yet specified". `mt create "…" --label
  wayfinder:map --no-place` (o map não disputa a fila de encaixe).
- **Child ticket**: uma issue com a label `wayfinder:<type>` (`research` /
  `prototype` / `grilling` / `task`) e, no topo do corpo, a linha
  `Part of <map-id>`. Não há hierarquia nativa — a linhagem é essa convenção de
  corpo + label, e o `mt list --label wayfinder:task` enumera os filhos.
- **Blocking**: depende do **nativo** do tracker — `mt dep add <child> <blocker>`
  (mesmo vault; `blocked_by` no frontmatter; `[blocked]` no `list`; ADR-0004).
  Uma ticket está **unblocked** quando todos os bloqueadores estão `done`.
  `mt ready` já pula blocked.
- **Frontier query**: `mt ready` (open + deferral passada + unblocked + sem claim)
  cruzado com `mt list --label wayfinder:task` e a linha `Part of` do corpo — a
  primeira na ordem do map vence.
- **Claim**: `mt claim <child-id> <sessão>` — contínua first write, **antes** de
  qualquer trabalho; a ticket claimada (`[claimed]` no `list`) sai de `ready` e é
  pulada por `pick-next` (ADR-0010). `mt unclaim <id>` devolve.
- **Resolve**: `mt comment <child-id> "<resposta>"`, depois `mt done <child-id>`,
  depois um `mt comment <map-id> "<context pointer>"` no map (a seção
  Decisions-so-far vive nos comments do map — append-only).

## Handoff entre skills

- `/triage` fala em *roles* — use as strings de `docs/agents/triage-labels.md`
  como labels do vault.
- `/to-tickets` publica um ticket por issue, blockers primeiro (ids reais antes
  de gravar as edges via `mt dep add`).
- `/to-spec` publica a spec no corpo de uma issue com `ready-for-agent`.
- `/implement-spec` faz `mt claim` no ticket antes de começar.
