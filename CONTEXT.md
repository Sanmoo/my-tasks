# my-tasks2

Um issue tracker pessoal baseado em arquivos: cada issue é um arquivo Markdown com frontmatter YAML, versionável em Git, operado por um CLI de binário único.

## Language

**Issue**:
Uma unidade de trabalho; um único arquivo Markdown com frontmatter YAML.
_Avoid_: task, tarefa, ticket

**Vault**:
Um diretório que contém as issues de um domínio de trabalho, junto de sua configuração. É a unidade de escopo de priorização e de `pick-next`.
_Avoid_: área, projeto, repositório

**Bookmark**:
Um apelido curto que aponta para um vault, definido em um arquivo de configuração detectado automaticamente. Usado como `@nome` nos comandos. Um dos bookmarks pode ser o favorito principal, usado quando nenhum `@` é informado.
_Avoid_: área, atalho, alias

**Rank**:
A posição de uma issue na ordem de prioridade dentro de um vault. Inteiro, único por vault, menor = primeiro. Issue sem rank pertence ao Backlog.
_Avoid_: prioridade, ordem (ordem é a sequência resultante; rank é o valor que a produz)

**Backlog**:
O conjunto de issues sem rank, abaixo da fila priorizada. É onde as ideias vivem até serem priorizadas.
_Avoid_: normal, não-priorizadas

**Deferred until**:
Data e hora a partir da qual uma issue fica disponível. Antes disso, a issue permanece `open` e continua **visível**, mas é **indisponível** para `pick-next` e `ready`. Quando `now >= deferred_until`, a Deferral está expirada e a issue volta a ficar disponível por conta própria.
_Avoid_: snooze, adiamento, defer-como-status

**Deferral expirada**:
Uma Deferred until cuja data/hora chegou (`now >= deferred_until`). É o sinal primário de `mt overdue` e o alvo do `mt undefer` em lote — o lembrete que o usuário agendou chegou, e `undefer` arquiva o lembrete limpando o campo.
_Avoid_: vencida, acordou, lembrete-como-estado

**Deadline**:
Data e hora limite de uma issue. Informativo; quando ultrapassado, a issue aparece em `overdue`.
_Avoid_: due date, prazo

**Blocked**:
Estado computável de uma issue: ela está blocked enquanto alguma issue listada no campo `blocked_by` (mesmo vault) não está `done`. Não é um status — não há transição nem operação de desbloqueio; fechar o bloqueador desbloqueia sozinho.
_Avoid_: status blocked, bloqueada

**Status**:
O estado de uma issue: `open`, `in_progress`, `done`, mais status personalizados definidos na configuração do vault. Não há máquina de estados imposta; apenas `pick-next` (→ `in_progress`, pulando issues blocked) e `done` (terminal) têm comportamento especial.
