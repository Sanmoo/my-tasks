# A aplicação de um plano de Rank é uma operação recuperável

Os planos de Rank — a renormalização do `prioritize`, os deslocamentos do
`top`/`bottom`/`rank`/`unrank`/`place`, o Encaixe concluído, a criação com
posição explícita — eram aplicados como **gravações independentes**: um laço
`Mutate` no CLI, uma gravação por Issue. Cada gravação era síncrona e íntegra
em si (o arquivo em disco era o antigo ou o mutado por completo), mas a
**operação** não era nada: uma falha na terceira de cinco gravações deixava a
Fila com Ranks duplicados — o estado intermediário era indistinguível da ordem
pedida. Na criação com `--top`/`--bottom`, os deslocamentos eram gravados
**antes** de a criação ser tentada, então uma criação que falhasse deixava
deslocamentos órfãos. O problema não é recalcular a Fila — o plano sempre foi
calculado centralmente, antes da primeira escrita; é **controlar o resultado
em disco** de uma operação que envolve várias Issues.

A aplicação passa a ser uma operação de persistência com preparação e
recuperação explícitas, dona de uma interface só, em `internal/issuefiles` —
o módulo que já era o único dono dos arquivos de Issue (ADR-0008). A
`ApplyRankPlan` recebe a intenção **já calculada** (`issue.Change` do domínio
e, opcionalmente, a Issue que a operação cria) e não decide nada sobre ordem,
Candidatas ou Ponto de encaixe — os cálculos ficam onde estavam, em
`internal/priority` e `internal/issue`, sem dependência nova na direção
proibida pelo ADR-0009. Os chamadores deixam de coordenar gravações
individuais e a sequência especial de criação com deslocamentos; o CLI
continua dono de argumentos, resolução do Vault, terminal, editor,
confirmações e apresentação de erros.

O contrato, escrito uma vez para todos os comandos de Rank
(`prioritize`, `top`, `bottom`, `rank`, `unrank`, `place`, `check --fix` e as
mudanças de Rank de `create`/`q`, Encaixe implícito incluído):

- **Preparar antes de escrever.** Todo alvo é validado (ID único path
  component, arquivo regular — a política do ADR-0008 intacta, symlinks
  recusados), seu conteúdo pretendido é renderizado e seus originais são lidos,
  **antes da primeira escrita**. Falha de preparação não muda nenhuma Issue —
  nem cria artefato.
- **Copiar os originais para `<Vault>/.mt-recovery/<op>/`** — junto do Vault,
  fora de `issues/`, logo fora da descoberta de Issues; nenhum listing lê a
  área, nenhum comando a interpreta. A cópia só existe enquanto a operação não
  estiver sabidamente completa.
- **Aplicar em ordem.** Uma gravação que retorna erro pode ter alterado o
  arquivo — o que conta para a recuperação é o que foi **tentado**, não o que
  retornou sucesso.
- **Recuperar em caso de falha**: todas as Issues tentadas voltam aos bytes e
  permissões originais — inclusive a cuja gravação falhou — e a Issue criada
  pela operação, se esta a criou, é removida; uma entrada preexistente que
  impediu a criação jamais é removida. Uma falha de restauração de uma Issue
  não impede a tentativa nas demais. Restauração completa não inclui inode,
  `mtime`, dono, ACLs — apenas **bytes e permissões**.
- **Ser honesto sobre o resultado.** Recuperação completa: erro diz que o
  estado anterior foi restaurado e nenhum artefato resta. Sucesso: idem.
  Recuperação incompleta: `RecoveryError` (falha operacional, exit 1) com o
  erro original, as Issues pendentes — distinguindo **remoção pendente**
  (Issue criada pela operação, sem original) de **restauração pendente** — e o
  **caminho absoluto** dos originais mantidos para recuperação manual. Nada
  afirma que a Fila está inalterada; nada executa recuperação automática
  depois; nenhuma regra de exclusão do Git é alterada.

O que o contrato **não** promete: recuperação após término abrupto (sem
journal durável, sem replay, nada crash-safe), proteção contra escritores
simultâneos, e que leituras concorrentes não observem o estado intermediário
durante a operação. Também não há garantia quando o filesystem impede a
própria restauração — nesse caso, diagnóstico e originais preservados.

Fora de escopo, de propósito: não é um mecanismo genérico de transações para
as demais mutações do tracker (`status`, `comment`, `claim` seguem pelo
`Mutate` de gravação única, cuja atomicidade de arquivo único já basta); não
há adapter novo de armazenamento nem interface intercambiável — a indução de
falhas dos testes é um hook interno ao pacote, exposto só por
`export_test.go`, e os cenários de processo seguem no binário compilado
(Seam 1). `.mt-recovery` não é conceito de domínio: Issue, Vault, Fila, Rank,
Backlog, Encaixe e Comparação mantêm os sentidos que já têm.
