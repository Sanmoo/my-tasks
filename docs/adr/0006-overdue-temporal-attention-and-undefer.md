# `overdue` é atenção temporal e `undefer` é limpeza de campo

O ADR-0002 registrou que "não existe operação de 'undefer'": a Deferral é um
campo de data e a disponibilidade é computada do relógio. Isso permanece
verdadeiro para o **modelo de disponibilidade** (`ready`/`pick-next`). Mas
o uso diário revelou duas necessidades que o campo, sozinho, não atende:

- **lembrete** — quando `now >= deferred_until` a issue volta a ficar
  disponível sem nenhum evento visível; nada lembra o usuário de que a data
  que ele mesmo escolheu chegou;
- **arquivamento** — limpar o campo (a issue foi lembrada e quer voltar ao
  convívio normal) exigia `mt edit` + cirurgia manual no frontmatter.

A virada, decidida em sessão de grilling (grill-with-docs, Q1–Q11):

- `mt overdue` deixa de ser só "Deadline estourado" e vira o **comando de
  atenção temporal** do Vault: Deferrais expiradas primeiro (sufixo
  `[expirada MM-DD]` — o sinal primário), depois Deadlines estourados
  (`[deadline MM-DD]`), cada grupo na ordem de Rank, uma issue com os dois
  sinais aparece uma única vez no grupo das expiradas, `done` fica de fora,
  blocked entra, vazio = silêncio com exit 0;
- `mt undefer` arquiva o lembrete: sem argumentos, limpa `deferred_until`
  de todas as Deferrais expiradas do Vault, imprimindo
  `Undeferred <id> (was <valor>)` por issue (zero expiradas = silêncio,
  exit 0 — o hábito diário roda às cegas); com um ID, limpa uma específica
  mesmo com Deferral ainda futura — o usuário mudou de ideia; uma issue sem
  o campo falha com exit 1.

`undefer` toca **só** o campo `deferred_until`: Status e Rank ficam
intocados, e o comando nunca tem opinião sobre prioridade ou estado. Sem
prompt de confirmação no lote: uma Deferral expirada não tem mais efeito
funcional (a disponibilidade já voltou), então limpá-la não perde estado
que importa.

A frase "não existe undefer" do ADR-0002 é superada **apenas** no sentido
de que a limpeza do campo agora tem um comando dedicado; o modelo de
disponibilidade dele — Deferral como campo, disponibilidade computada do
relógio, `ready`/`pick-next` intocados — permanece intacto.
