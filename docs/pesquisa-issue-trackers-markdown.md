# Pesquisa: issue/task trackers CLI com Markdown e YAML frontmatter

## Resumo executivo

Existe hoje um candidato que satisfaz **todos os requisitos funcionais declarados**: **nd**. Ele mantém uma tarefa por arquivo Markdown com YAML frontmatter, oferece CRUD, `defer --until`, filtros, os cinco estados requeridos, estados adicionais configuráveis, grafo de dependências/ready queue e modos locais compatíveis com Git. A ressalva é de **risco de adoção**, não de cobertura funcional: trata-se de um projeto novo e menos estabelecido que Taskwarrior/Taskell.

**Crumbs** é a alternativa mais próxima e simples, mas seus estados são fixos. **taskmd** é forte em arquivos, CRUD, filtros e dependências, porém não documenta snooze/defer com despertar futuro nem estados customizáveis. Ferramentas conhecidas como Taskell, Taskwarrior, Todoman e Taskbook falham no requisito estrutural de “uma issue Markdown com YAML frontmatter”.

## Critérios usados

- **Formato estrito:** uma tarefa/issue em arquivo `.md`, com seus campos estruturados em YAML frontmatter (não apenas checkboxes ou um Markdown agregado).
- **CRUD:** criar, ler/listar, atualizar e remover/arquivar.
- **Futuro:** defer/snooze com data de reativação, e não apenas `due`.
- **Workflow:** `open`, `in-progress`, `blocked`, `closed` e estados customizáveis.
- **Grafo:** dependências/relações e, idealmente, cálculo de trabalho desbloqueado.
- **Local/Git:** sem serviço obrigatório; arquivos legíveis, diffáveis e versionáveis.

## Tabela comparativa

Legenda: ✅ atende; ◐ parcial; ❌ não atende; ? não confirmado na documentação primária.

| Ferramenta | Markdown + YAML por item | CRUD | Defer/snooze futuro | Filtros | Estados requeridos | Estados customizáveis | Relações/deps | Local/Git | Veredito |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| **nd** | ✅ | ✅ | ✅ `defer --until` | ✅ | ✅ | ✅ | ✅ grafo/ready/ciclos | ✅ | **Atende tudo** |
| **Crumbs** | ✅ | ✅ | ✅ `defer --until` | ✅ | ✅ | ❌ fixos | ✅ | ✅ | Melhor alternativa simples; falta customização |
| **taskmd** | ✅ | ✅ | ❌ não documentado | ✅ | ✅ (nomes próximos) | ❌ enum definido | ✅ grafo/next | ✅ | Forte, mas sem snooze e workflow customizável |
| **kanban-md** | ✅ | ✅ | ◐ `due`, sem snooze confirmado | ✅ | ✅ configuráveis | ✅ | ✅ | ✅ | Muito próximo; falta defer temporal comprovado |
| **git-issues** | ✅ | ✅ | ❌ | ✅ | ◐ open/closed | ❌ | ✅ `blocks` | ✅ | Issue tracker enxuto, workflow insuficiente |
| **issuectl** | ✅ | ✅ | ? | ✅ consultas/bulk | ◐ | ? | ✅ `related`, `blocked_by` | ✅ merge driver | Próximo do Beads, mas defer/custom status não comprovados |
| **Taskell** | ❌ arquivo agregado | ✅ via TUI | ◐ due date | ◐ | ◐ colunas | ◐ colunas | ◐ subtarefas | ✅ | Não cumpre formato por issue/frontmatter |
| **Taskwarrior** | ❌ SQLite/TaskChampion | ✅ | ✅ `wait` | ✅ muito fortes | ◐ estados fixos | ◐ UDA não equivale a estados | ✅ | ◐ local, mas não Markdown/Git | Funcionalmente forte; formato incompatível |
| **Todoman** | ❌ iCalendar `.ics` | ✅ | ◐ start/due | ✅ | ◐ iCalendar | ❌ | ❌ | ◐ arquivos locais | Formato incompatível |
| **Taskbook** | ❌ JSON | ✅ | ❌ | ✅ | ◐ pending/started/done | ❌ | ❌ | ◐ local | Formato incompatível |
| **Imdone (clássico)** | ❌ extrai TODOs/checklists | ◐ | ? | ✅ | ◐ listas | ✅ listas | ❌/limitado | ✅ | Não confundir com issue por frontmatter |
| **Mark** | ❌ | ❌ para tracking | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | Publicador Markdown; frontmatter ainda é pedido de feature |

## Achados detalhados

### 1. nd é o único substituto pronto que cobre toda a matriz

O repositório oficial define o **nd** como tracker Git-native sem servidor ou banco, com uma issue por arquivo Markdown/Obsidian e YAML frontmatter. O CLI acrescenta CRUD, busca, filtros, soft-delete, grafo de dependências, cálculo de `ready`, bloqueios e ciclos. Traz nativamente `open`, `in_progress`, `blocked`, `deferred`, `closed`, e permite estados adicionais e uma máquina de estados configurável em `.nd.yaml`. O comando `nd defer <id> --until=YYYY-MM-DD` cobre precisamente snooze/data futura. Há modo `--track-issues` para versionar os arquivos diretamente e um fluxo de branch de backlog/sync para múltiplos agentes. [README oficial](https://github.com/paivot-ai/nd/blob/main/README.md)

Exemplo de configuração oficial:

```yaml
status_custom: "review,qa,rejected"
status_sequence: "open,in_progress,review,qa,closed"
status_fsm: true
```

**Avaliação:** substituto funcional pronto mais próximo do Beads, inclusive pela ready queue baseada em dependências. Antes de adoção ampla, convém validar release cadence, estabilidade do formato e comportamento de merge/worktrees num piloto.

### 2. Crumbs atende quase tudo, mas fixa o workflow

O **Crumbs** usa um `.md` por item com YAML frontmatter e cache CSV reconstruível. Documenta create/list/show/edit, delete/clean, filtros por status/tag/prioridade/tipo/fase, dependências e `next` que ignora bloqueados e itens deferidos até sua data. Também implementa `crumbs defer ID --until DATA`. Entretanto, o schema enumera somente `open | in_progress | blocked | deferred | closed`; não há configuração oficial de estados extras. [Repositório oficial](https://github.com/evensolberg/crumbs)

**Avaliação:** excelente opção se estados customizáveis puderem ser removidos do escopo; caso contrário, exigiria fork/alteração.

### 3. taskmd é sólido e Git-first, mas “due” não é snooze

O **taskmd** especifica um arquivo Markdown por tarefa, YAML frontmatter, CRUD (`add`, `get`, `set`, `rm`, `archive`), filtros, busca, board, dependências, grafo e comando `next`. O enum inclui `pending`, `in-progress`, `completed`, `in-review`, `blocked`, `cancelled`. A especificação define dependências e datas terminais; datas `due` aparecem para fases, mas não foi encontrada operação equivalente a “ocultar até X e reabrir/ficar acionável em X”. Os estados são um enum da especificação, não uma lista configurável por projeto. [Repositório oficial](https://github.com/driangle/taskmd) · [Especificação oficial](https://github.com/driangle/taskmd/blob/main/docs/taskmd_specification.md) · [Releases/CLI](https://github.com/driangle/taskmd/releases)

**Avaliação:** bom protocolo e boa CLI, mas não satisfaz defer temporal nem estados arbitrários sem extensão.

### 4. kanban-md é promissor, porém falta evidência de snooze

O **kanban-md** mantém um arquivo Markdown com frontmatter por tarefa; seu `config.yml` permite definir colunas/estados como `backlog`, `todo`, `in-progress`, `review`, `done`, além de WIP limits. A CLI documenta edição, due date, parent, dependências, block/unblock e claim/release. Não foi encontrada, na documentação oficial consultada, semântica de `defer_until` que retire a tarefa da fila e a torne acionável no futuro. [Repositório oficial](https://github.com/antopolskiy/kanban-md)

**Avaliação:** candidato a piloto secundário; `due` sozinho não satisfaz snooze.

### 5. issuectl e git-issues são próximos do modelo Beads, mas incompletos para este contrato

O **issuectl** guarda `issues/<id>/item.md`, tem campos `related` e `blocked_by`, consultas, bulk update, JSON e merge driver específico para frontmatter — características excelentes para Git e agentes. Porém, nas fontes oficiais localizadas, não ficaram comprovados defer com data e estados arbitrários. [Repositório oficial](https://github.com/jarimustonen/issuectl)

O **git-issues** guarda cada issue em `.issues/NNNN-slug.md`, lista com filtros e suporta relações `blocks`, mas seu fluxo documentado é essencialmente open/close e não cobre snooze/custom workflow. [Repositório oficial](https://github.com/steviee/git-issues)

### 6. Candidatos famosos foram excluídos por formato, não por falta de qualidade

- **Taskell:** armazena um board inteiro em `taskell.md` usando headings/listas; tem subtarefas e due dates, mas não uma issue por arquivo com YAML frontmatter. O repositório está arquivado. [Repositório oficial](https://github.com/smallhadroncollider/taskell)
- **Taskwarrior:** oferece `wait`, filtros sofisticados e `depends`, mas a versão atual armazena em `taskchampion.sqlite3`; exportação JSON/YAML não transforma Markdown em source of truth. [Manual oficial](https://taskwarrior.org/docs/man/task.1/) · [Modelo oficial](https://taskwarrior.org/docs/task/)
- **Todoman:** usa arquivos individuais, porém no padrão iCalendar `.ics`, não Markdown/YAML. [Documentação oficial](https://todoman.readthedocs.io/en/stable/) · [Repositório oficial](https://github.com/pimutils/todoman/)
- **Taskbook:** possui CRUD básico, begin/check e filtros, mas persiste JSON em `~/.taskbook/storage`. [Repositório oficial](https://github.com/klaudiosinani/taskbook)
- **Imdone clássico:** encontra tags TODO/FIXME e tarefas estilo todo.txt em Markdown/código; isso não equivale a um documento por issue com schema YAML. [Repositório oficial](https://github.com/imdone/imdone-core)
- **Mark:** é ferramenta de publicação de Markdown; suporte a YAML frontmatter como metadata ainda aparece como solicitação de recurso, portanto não é issue tracker relevante. [Issue oficial](https://github.com/kovetskiy/mark/issues/792)

## Comparação com Beads

Beads continua mais maduro no conceito agent-first: grafo, `ready`, claim atômico, múltiplos tipos de relação, defer e filtros. Contudo, sua fonte de verdade atual é Dolt; o JSONL é intercâmbio, não Markdown por issue. [Repositório oficial](https://github.com/gastownhall/beads) · [`bd ready` oficial](https://gastownhall.github.io/beads/cli-reference/ready)

O **nd** é a tradução mais direta desse modelo para Markdown/frontmatter: grafo e ready queue semelhantes, estados requeridos/customizáveis e defer temporal, preservando arquivos humanos. Crumbs replica o núcleo, mas não a extensibilidade do workflow.

## Conclusão e recomendação build-vs-adopt

**Conclusão clara:** sim, existe um substituto funcional pronto: **adotar `nd` em piloto**. Ele é o único candidato encontrado que, segundo a documentação primária, satisfaz simultaneamente todos os requisitos.

Recomendação:

1. **Adopt/pilot:** testar `nd` num repositório real por 1–2 ciclos, com `--track-issues` ou branch de backlog conforme o modelo de colaboração.
2. Validar no piloto: merges concorrentes, recuperação/sync, estabilidade do schema, performance com backlog representativo, comportamento exato de wake-up após `--until` e automação JSON.
3. **Fallback:** Crumbs se workflow fixo for aceitável; taskmd se snooze puder ser implementado externamente.
4. **Build somente se** o piloto do nd falhar em manutenção/estabilidade ou se for obrigatório controlar integralmente formato e semântica. Nesse caso, construir sobre um schema mínimo semelhante ao nd/Crumbs é preferível a adaptar Taskwarrior/Todoman, pois estes violam o requisito central de armazenamento.

## Lacunas e riscos residuais

- A pesquisa documental não substitui teste de instalação e uso; nenhum binário foi executado.
- Projetos muito novos podem mudar schema/CLI rapidamente; stars, número de usuários e política de compatibilidade não provam estabilidade.
- Não foi confirmado por teste se `nd defer --until` reabre automaticamente o status ou apenas volta a incluir o item nas consultas acionáveis; a documentação confirma defer com data, mas esse detalhe deve entrar no piloto.
- Não se verificou empiricamente resolução de conflitos quando duas branches alteram o mesmo arquivo.
- issuectl e kanban-md merecem reavaliação se suas documentações adicionarem defer temporal explícito.

## Fontes mantidas e descartadas

**Mantidas (primárias):** repositórios/documentação oficiais de [nd](https://github.com/paivot-ai/nd/blob/main/README.md), [Crumbs](https://github.com/evensolberg/crumbs), [taskmd](https://github.com/driangle/taskmd), [kanban-md](https://github.com/antopolskiy/kanban-md), [issuectl](https://github.com/jarimustonen/issuectl), [git-issues](https://github.com/steviee/git-issues), [Taskell](https://github.com/smallhadroncollider/taskell), [Taskwarrior](https://taskwarrior.org/docs/man/task.1/), [Todoman](https://github.com/pimutils/todoman/), [Taskbook](https://github.com/klaudiosinani/taskbook), [Imdone](https://github.com/imdone/imdone-core), [Mark](https://github.com/kovetskiy/mark/issues/792) e [Beads](https://github.com/gastownhall/beads).

**Descartadas:** posts agregadores, listas SEO e comentários de terceiros; forks/ports redundantes; ferramentas que apenas gerenciam checkboxes sem metadata por item. Esses materiais não oferecem evidência tão forte quanto README, especificação, código e manuais oficiais.
