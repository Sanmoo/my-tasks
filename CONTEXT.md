# my-tasks2

Um issue tracker pessoal baseado em arquivos: cada issue é um arquivo Markdown com frontmatter YAML, versionável em Git, operado por um CLI de binário único.

## Language

**Issue**:
Uma unidade de trabalho; um único arquivo Markdown com frontmatter YAML.
Em disco, a Issue é sempre um arquivo regular dentro de `issues/`: um
symlink com nome de Issue não é uma Issue — é pulado pela listagem e
recusado na leitura, na escrita, na edição e na criação.
_Avoid_: task, tarefa, ticket

**Vault**:
Um diretório que contém as issues de um domínio de trabalho, junto de sua configuração. É a unidade de escopo de priorização e de `pick-next`.
_Avoid_: área, projeto, repositório

**Bookmark**:
Um apelido curto que aponta para um vault, definido em um arquivo de configuração detectado automaticamente. Usado como `@nome` nos comandos. Um dos bookmarks pode ser o favorito principal, usado quando nenhum `@` é informado.
_Avoid_: área, atalho, alias

**ID prefix**:
A parte de uma issue ID antes do primeiro `-` (ex.: `dom` em `dom-xyz`); identifica o vault a que a Issue pertence: nos comandos que recebem uma key, quando nenhum `@bookmark` ou `--vault` é dado, o vault é resolvido pelo prefixo antes de recorrer ao bookmark default.
_Avoid_: prefixo-do-vault

**Rank**:
A posição de uma issue na ordem de prioridade dentro de um vault. Inteiro, único por vault, menor = primeiro. Issue sem rank pertence ao Backlog.
_Avoid_: prioridade, ordem (ordem é a sequência resultante; rank é o valor que a produz)

**Backlog**:
O conjunto de issues sem rank, abaixo da fila priorizada. É onde as ideias vivem até serem priorizadas.
_Avoid_: normal, não-priorizadas

**Encaixe**:
A operação de descobrir o Rank de uma Issue comparando-a com as Issues já na fila, uma de cada vez. É o que `mt place` faz e o que `mt create`/`mt q` fazem por padrão.
_Avoid_: priorizar, ordenar, inserir

**Comparação**:
Uma pergunta estrita de uma sessão de Encaixe: entre duas Issues, qual vem antes. Sem empate — Rank é único por vault. A resposta "indiferente" não é um empate: é a ordem já gravada que decide (a Issue encaixada entra logo após a Candidata).
_Avoid_: pergunta, duelo, voto

**Candidata**:
A Issue da fila que ocupa o meio do intervalo corrente de um Encaixe e é o segundo item da Comparação. Issue no Backlog e Issue não-priorizável (`done`, status custom) nunca são Candidatas.
_Avoid_: oponente, referência, pivô

**Ponto de encaixe**:
O Rank que uma sessão de Encaixe conclui para a Issue encaixada.
_Avoid_: posição (posição é o argumento explícito de `mt rank <id> <n>`)

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

**Comment**:
Um bloco de texto datado anexado ao fim do corpo da Issue, preservado para sempre (append-only), identificado dentro da Issue por uma âncora estável. Não é campo de frontmatter nem status: uma transição pode carregar um Comment, mas ele é escrito no corpo, na mesma gravação — nenhum metadado guarda o texto.
_Avoid_: nota, anotação, observação, motivo de fechamento, close reason
