# mt — issue tracker pessoal, git-friendly

`mt` é um issue tracker pessoal baseado em arquivos: cada **Issue** é um
arquivo Markdown com frontmatter YAML, versionável em Git, operado por um
CLI de binário único (Go, sem runtime externo). Um **Vault** — um diretório
com as Issues de um domínio e sua configuração — é a unidade de escopo de
priorização e de `pick-next`. Tudo é legível por humanos e por LLMs, e o Git
é o histórico.

Para o vocabulário exato do domínio (Issue, Vault, Bookmark, Rank, Backlog,
Deferred until, Deadline, Status), veja [`CONTEXT.md`](CONTEXT.md).

## Instalação

### Binário único a partir do fonte

```sh
make build                       # compila bin/mt
install -m 0755 bin/mt ~/.local/bin/mt   # ou qualquer diretório no PATH
```

O binário é único e não exige runtime externo; é só copiá-lo para outra
máquina.

### Via mise (como o nd)

Com um tag de versão no repositório (ex.: `v1.0.0`), o `mt` entra no mise
pelo backend `go:` — a mesma entrada única do `nd` no config global:

```toml
# ~/.config/mise/config.toml
[tools]
"go:github.com/Sanmoo/my-tasks2/cmd/mt" = "latest"
```

```sh
mise install
```

O mise baixa, compila e cria o shim de `mt`. Sem tags, o backend `go:` não
resolve versão — nesse caso use a instalação a partir do fonte acima (ou
`go install github.com/Sanmoo/my-tasks2/cmd/mt@latest` após o primeiro
tag).

## Começo rápido

```sh
# 1. Cria um vault com prefixo de ID derivado do nome do diretório
mt init ~/dev/pkm/.vault

# 2. Registra o bookmark e o define como padrão (config global)
mt bookmark add pkm ~/dev/pkm/.vault
printf 'default: pkm\nbookmarks:\n  pkm: ~/dev/pkm/.vault\n' \
  > ~/.config/mt/config.yaml

# 3. A partir daqui, os comandos funcionam sem endereçar nada:
#    o bookmark padrão resolve o vault.
mt create "comprar material"
mt q "ideia rápida"                  # imprime só o ID
mt list
```

## Endereçamento de vault

Comandos que operam sobre um vault o resolvem nesta ordem de precedência:

1. `@bookmark` — o token `@nome` pode aparecer em qualquer posição dos args
   (ex.: `mt list @pkm`);
2. `--vault <path>` — flag global de qualquer comando (ex.: `mt list --vault ~/dev/pkm/.vault`);
3. `default` — o bookmark padrão da config global.

Sem nenhum dos três, o comando falha com instruções no stderr (exit 1). Não
há detecção por diretório corrente: o vault é sempre endereçado
explicitamente.

## Comandos

Resumo:

| Comando | O que faz |
| --- | --- |
| `mt init [dir]` | cria um Vault (`issues/` + `mt.yaml`) |
| `mt create <título>` | cria uma Issue |
| `mt q <título>` | cria uma Issue e imprime só o ID |
| `mt show <id>` | mostra a Issue completa (frontmatter + corpo) |
| `mt edit <id>` | abre a Issue no `$EDITOR` |
| `mt done <id>` (alias `close`) | fecha a Issue (carimba `completed_at`) |
| `mt reopen <id>` | reabre (limpa `completed_at` e `started_at`) |
| `mt status <id> <status>` | transição livre de status |
| `mt defer <id> <quando>` | adia a Issue até uma data/hora |
| `mt dep add <id> <bloqueador>` / `mt dep rm <id> <bloqueador>` | registra/remove dependência (`blocked_by`) |
| `mt comment <id> <texto>` | anexa um comentário com timestamp |
| `mt list` | lista na ordem de prioridade |
| `mt ready` | lista as Issues disponíveis agora |
| `mt overdue` | lista as Issues com Deadline estourado |
| `mt pick-next` | inicia a próxima Issue disponível |
| `mt prioritize` | prioriza no `$EDITOR` (fila × Backlog) |
| `mt top <id>` / `mt bottom <id>` | move para a primeira/última posição da fila |
| `mt rank <id> <n>` | insere na posição `n` da fila |
| `mt unrank <id>` | devolve a Issue ao Backlog |
| `mt check [--fix]` | audita a integridade do Vault |
| `mt bookmark add/list/rm` | gerencia a config global |
| `mt help [comando]` | ajuda de qualquer comando |

### `mt init [dir]`

Cria um vault em `dir` (padrão: diretório corrente): o diretório `issues/`
e o arquivo de configuração `mt.yaml` com o prefixo de ID e a lista de
status. Recusa sobrescrever um vault existente.

```sh
mt init ~/dev/pkm/.vault
# → Vault ready at /home/sanmoo/dev/pkm/.vault
```

- `--prefix <p>` — prefixo de ID (padrão: derivado do nome do diretório,
  minúsculo, sem não-alfanuméricos, ≤ 8 caracteres);
- `--status <s>` — status do vault, repetível (padrão: `open`,
  `in_progress`, `done`).

### `mt create <título>` e `mt q <título>`

Escrevem uma nova Issue (`issues/<id>.md`) com o schema da spec: `title`,
`status: open`, `labels` e `created_at` no frontmatter, e o corpo vazio
(`## Description`, `## Notes`, `## Comments`). O ID é o prefixo do vault
mais um sufixo aleatório curto (`pkm-055`); `created_at` é carimbado
automaticamente. O título é a junção dos argumentos restantes — não precisa
de aspas no shell.

```sh
mt create "[Niver Edu] comprar material de Assaí"
# → Created pkm-055

mt create "ideia para depois" --label compras
# → Created pkm-5qa8

mt q "anotar rápido"
# → pkm-07r0        (só o ID — para capturar sem sair do fluxo)
```

- `--label <l>` — label livre, repetível (create).

### `mt show <id>` e `mt edit <id>`

`show` imprime o arquivo da Issue exatamente como está — frontmatter, corpo
e comentários. `edit` abre o arquivo no `$EDITOR` (split em whitespace, o
caminho nunca passa por shell); qualquer edição é preservada, inclusive
edições manuais do frontmatter.

```sh
mt show pkm-055
mt edit pkm-055        # exige $EDITOR configurado
```

### Transições de status: `done`, `close`, `reopen`, `status`

```sh
mt done pkm-055        # status done + completed_at carimbado; close é alias
mt reopen pkm-055      # open, limpando completed_at e started_at
mt status pkm-055 in_progress   # transição livre, validada contra os status do vault
```

Não há máquina de estados: qualquer status da lista do vault é alcançável
com `status`. Só `done` (terminal, carimba `completed_at`) e `pick-next`
(→ `in_progress`) têm comportamento especial. Status fora da lista do vault
é rejeitado na hora (exit 1) e reportado por `check`.

### `mt defer <id> <quando>`

Adia a Issue até `deferred_until`, deixando-a `open` mas indisponível — ela
volta a ficar disponível sozinha quando `now >= deferred_until`; não existe
"undefer". O horário é preservado.

```sh
mt defer pkm-055 "26-08-20 08:00"    # absoluto: YY-MM-DD HH:MM (hora importa)
mt defer pkm-055 +2d                  # relativo: +2d, +1w, +3h
# → pkm-055 deferred until 2026-08-20T08:00
```

### `mt dep add <id> <bloqueador>` | `mt dep rm <id> <bloqueador>`

Registra dependências no campo `blocked_by` (mesmo vault, direção única: a
Issue registra quem a bloqueia). Uma Issue está **blocked** enquanto alguma
Issue listada não está `done` — estado computado, não status: não há
transição nem operação de desbloqueio; fechar o bloqueador desbloqueia
sozinho, e reabri-lo rebloqueia.

```sh
mt dep add pkm-002 pkm-001   # pkm-001 bloqueia pkm-002
# → pkm-002 is now blocked by pkm-001

mt dep rm pkm-002 pkm-001
# → pkm-002 is no longer blocked by pkm-001
```

- `dep add` exige que o bloqueador exista no vault e que não seja a própria
  Issue (erros de usuário, exit 1); argumentos malformados são erro de uso
  (exit 2);
- `dep rm` é idempotente — remove referências órfãs (ex.: para uma Issue
  apagada) sem reclamar;
- Issues bloqueadas aparecem com sufixo `[blocked]` no `list` e são puladas
  por `ready` e `pick-next`;
- `mt check` valida as referências: existência no vault, sem auto-bloqueio,
  sem ciclos.

### `mt comment <id> <texto>`

Anexa um comentário à seção `## Comments` da Issue: heading com timestamp,
o texto e uma âncora estável (`<!-- comment: … -->`) por comentário.
Append-only: o corpo existente é preservado byte a byte.

```sh
mt comment pkm-055 "comprei metade da lista"
```

```markdown
## Comments
### 2026-08-16T14:05
Comprei metade da lista.
<!-- comment: 4f2b9c1a -->
```

### `mt list`

Lista as Issues na ordem de prioridade: fila (menor Rank primeiro), depois
o Backlog (sem Rank, por `created_at`), com ID como desempate final. Cada
linha é um glyph de status, o ID e o título:

```text
○ pkm-001  primeira
◐ pkm-002  em andamento
○ pkm-003  ideia do backlog
```

Issues `done` e adiadas para o futuro ficam ocultas por padrão. Issues
bloqueadas (algum ID de `blocked_by` não está `done`) ficam visíveis com
sufixo `[blocked]` — com `--all`, o sufixo de adiamento e o `[blocked]`
aparecem juntos. Flags:

- `--all` — mostra `done` e adiadas; adiadas ganham sufixo `[defer MM-DD HH:MM]`;
- `--status <s>` — filtra por status;
- `--label <l>` — filtra por label, repetível.

Rank duplicado (edição manual) gera `Warning: duplicate rank: 1` no stderr,
antes da listagem. Os glyphs são `○` (open), `◐` (in_progress), `●` (done)
e `?` para status customizados.

### `mt ready` e `mt overdue`

- `ready` — as Issues `open` e disponíveis agora (`now >= deferred_until` e
  nenhum bloqueador não-`done`), na ordem de prioridade de `list`;
- `overdue` — as Issues não-`done` com `deadline` no passado. O Deadline é
  informativo: não bloqueia nada, só aparece aqui.

Ambas respeitam o formato de linha de `list`; sem correspondências, a saída
é vazia com exit 0.

### `mt pick-next`

Inicia a próxima Issue disponível: a `open` de menor Rank; sem Issues
rankeadas, a mais antiga do Backlog (desempate por ID). Adiadas para o
futuro e bloqueadas (algum bloqueador não-`done`) são puladas. A Issue
escolhida vira `in_progress` e recebe `started_at`. Várias Issues
`in_progress` simultâneas são permitidas.

```sh
mt pick-next
# → pkm-002 is now in_progress
```

Sem nada disponível: mensagem clara no stderr e exit 1. Rank duplicado no
vault: recusa com erro (a ambiguidade nunca é resolvida no chute).

### `mt prioritize`

Abre o `$EDITOR` com um buffer das Issues `open` e `in_progress`:

```text
# Edit ranking for this vault
# Reorder lines. Use [P] for prioritized, [ ] for backlog.
# Do not edit issue IDs. Save and close to continue.

[P] pkm-055  <título>
[P] pkm-07r0  <título>
[ ] pkm-5qa8  <título>
```

- Reordenar linhas `[P]` muda a ordem da fila;
- `[ ]` ↔ `[P]` move a Issue entre o Backlog e a fila;
- Ao salvar, os ranks são renormalizados 1..N e só os arquivos cujo rank
  mudou são reescritos (apply in-process, sem subprocesso por item);
- Buffer inválido (ID desconhecido/duplicado, linha malformada, Issue não
  priorizável — `done`/status customizado — no buffer) → erro, nada é
  aplicado (exit 1).

### Ajustes rápidos: `top`, `bottom`, `rank`, `unrank`

```sh
mt top pkm-055          # primeira posição da fila
mt bottom pkm-055       # última posição da fila
mt rank pkm-055 3       # insere na posição 3 (1-based)
mt unrank pkm-055       # de volta ao Backlog
```

Como no `prioritize`, a fila é renormalizada e só os arquivos alterados são
reescritos. Posição fora da fila atual → erro (exit 1); posição que não é
inteiro positivo → erro de uso (exit 2).

### `mt check [--fix]`

Audita a integridade do vault:

- Rank duplicado — **erro** (exit 1), com os IDs envolvidos;
- Lacuna de Rank — **aviso** no stderr (`Warning: rank gap: 2`), não é erro:
  a fila continua não-ambígua;
- YAML/frontmatter malformado — erro, nomeando a Issue;
- Status fora da lista configurada do vault — erro;
- Datetime em formato inválido — erro;
- `blocked_by` — referência a Issue inexistente, auto-bloqueio ou ciclo —
  erro, nomeando os IDs envolvidos.

`--fix` renormaliza os Ranks para 1..N (escrevendo só os arquivos alterados)
e revalida. Vault íntegro: `OK` no stdout, exit 0.

```sh
mt check --vault ~/dev/pkm/.vault
# → OK
```

### `mt bookmark add <nome> <caminho>` | `list` | `rm <nome>`

Gerencia a config global (ver [Configuração global](#configuração-global)).
Nomes são sem `@` (letras, dígitos, `-` e `_`); o `@nome` é a forma de
endereçar nos outros comandos. `rm` do bookmark padrão também limpa o
padrão.

```sh
mt bookmark add pkm ~/dev/pkm/.vault
# → Bookmark pkm added.

mt bookmark list
# → pkm -> ~/dev/pkm/.vault (default)

mt bookmark rm pkm
# → Bookmark pkm removed.
```

O bookmark padrão é definido pela chave `default:` da config global.

### `mt help [comando]`

`mt help` (ou `--help`, ou `mt` sem argumentos) mostra a ajuda raiz;
`mt help <comando>` mostra a ajuda do comando. Tópico desconhecido é erro de
uso (exit 2). O comando auxiliar `mt completion <shell>` (adicionado pelo
Cobra) gera scripts de completação para o shell.

## Configuração global

Detectada automaticamente: `$XDG_CONFIG_HOME/mt/config.yaml` ou, sem
`XDG_CONFIG_HOME`, `~/.config/mt/config.yaml`.

```yaml
default: bjd
bookmarks:
  bjd: ~/dev/github.com/Sanmoo/pkm/.vault
  dom: ~/dev/github.com/Sanmoo/dom/.vault
```

- `bookmarks` — mapa `nome → caminho` do vault. O caminho pode começar com
  `~`, expandido na resolução;
- `default` — nome do bookmark usado quando nenhum `@bookmark` ou `--vault`
  é informado.

Gerenciada pelos comandos `mt bookmark add/list/rm`, ou editada à mão.

## Configuração do vault

`mt.yaml` na raiz do vault, criado por `mt init`:

```yaml
prefix: pkm
status: [open, in_progress, done]
```

- `prefix` — prefixo dos IDs das Issues (`pkm-055`). Sem prefixo, `create`
  falha com instruções;
- `status` — lista de status do vault. Vazia/ausente = os padrões
  `open, in_progress, done`. Valida `mt status <id> <status>` e é a lista de
  referência do `mt check`. Status customizados aparecem com glyph `?` no
  `list`.

## Schema da Issue

Uma Issue é um arquivo `issues/<id>.md`:

```markdown
---
title: "[Niver Edu] comprar material de Assaí conforme doc compartilhado"
status: open
labels: [compras, familia]
created_at: 2026-08-15T09:30
rank: 2
deferred_until: 2026-08-20T08:00
deadline: 2026-08-22T18:00
blocked_by: [pkm-042]
---

## Description
## Notes
## Comments
### 2026-08-16T14:05
Comprei metade da lista.
<!-- comment: 4f2b9c1a -->
```

Regras de campo:

- Sempre presentes: `title`, `status`, `labels`, `created_at`;
- Só quando têm valor: `rank`, `deferred_until`, `deadline`, `started_at`,
  `completed_at`, `blocked_by`;
- Sem `id` (o nome do arquivo é a autoridade) e sem `updated_at` (o Git é o
  histórico);
- Datas são `YYYY-MM-DDTHH:MM` naive (sem timezone, sem segundos) — diffs
  mínimos e leitura humana;
- Corpo com apenas `## Description`, `## Notes`, `## Comments`;
- Comentário = heading de timestamp + âncora estável `<!-- comment: <curto> -->`.

## Exit codes e streams

Convenção de saída do processo — a mesma para todos os comandos:

| Código | Significado | Exemplos |
| --- | --- | --- |
| `0` | sucesso | qualquer comando que cumpriu o que pediu |
| `1` | erro de usuário — comando bem-formado que falhou contra o estado atual | vault indefinido/bookmark desconhecido, Issue não encontrada, status fora da lista do vault, nada disponível em `pick-next`, rank duplicado, edição inválida no `prioritize`, `init` num vault existente |
| `2` | erro de uso — invocação malformada | comando/flag/tópico de help desconhecido, contagem de argumentos errada, argumento malformado (nome de bookmark inválido, posição de rank não inteiro positivo, ID de Issue com separador de caminho, tempo de `defer` não-parseável, dois `@bookmark`) |

Regras de streams:

- Resultados vão para o **stdout**; erros e avisos vão para o **stderr**;
- Todo erro é impresso como `Error: <mensagem>` no stderr;
- Erros de uso ganham a dica `Run 'mt --help' for usage.` em seguida;
- Avisos (rank duplicado no `list`, lacuna de rank no `check`) também vão
  para o stderr — nunca poluem o stdout.

Essa convenção é coberta por testes de comportamento no e2e
(`e2e/features/smoke.feature` e cenários por comando) e pode ser auditada de
uma vez com `make audit`.

## Desenvolvimento

```sh
make check    # unit + e2e + coverage gate + mutation — o alvo único
make audit    # sonda exit code e stderr de todos os comandos no binário
make build    # bin/mt
```

Testes e seams em [`TESTING.md`](TESTING.md).
