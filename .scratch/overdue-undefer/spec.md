Status: ready-for-agent

# Spec: `mt overdue` como atenção temporal + `mt undefer`

## Problem Statement

O usuário usa `mt overdue` diariamente como sua checagem do que o tempo cobra. Hoje o comando só sabe de uma relação com o tempo: Deadline estourado. Mas a relação que o usuário considera mais importante é outra — **a Deferral cuja hora chegou**. Adiar uma Issue é agendar uma volta: quando `now >= deferred_until` a Issue simplesmente fica disponível de novo (ADR-0002), sem nenhum evento visível — nada lembra o usuário de que a data que ele mesmo escolheu chegou. E quando o usuário quer arquivar o lembrete (limpar o campo `deferred_until` de uma Issue que ele já relembrou e quer de volta no convívio normal), o único caminho hoje é `mt edit` + cirurgia manual no frontmatter.

## Solution

`mt overdue` passa a ser o **comando de atenção temporal** do Vault: Deferrais expiradas primeiro (o sinal primário), depois Deadlines estourados, cada grupo na ordem de Rank do Vault, e cada linha marcada com o motivo de estar ali. Novo comando `mt undefer` arquiva o lembrete: sem argumentos, limpa `deferred_until` de todas as Deferrais expiradas do Vault; com um ID, limpa uma Issue específica (mesmo uma Deferral ainda futura — mudou de ideia). `undefer` mexe **só no campo**: Status e Rank ficam intocados, e a semântica de disponibilidade de `ready`/`pick-next` (ADR-0002) não muda.

## User Stories

1. As a usuário que adia Issues, quero que `mt overdue` liste as Issues cuja Deferred until chegou, para ser lembrado de cada Issue adiada quando chegar a hora.
2. As a usuário, quero que as Deferrais expiradas apareçam **antes** dos Deadlines estourados, para que o sinal que mais me importa não se perca no meio.
3. As a usuário, quero que cada linha marcada indique o motivo (`[expirada MM-DD]` vs `[deadline MM-DD]`), para distinguir "chegou a hora que eu escolhi" de "estou atrasado".
4. As a usuário, quero que uma Issue com Deferred until expirada **e** Deadline estourado apareça uma única vez, no grupo das expiradas, para a vista não duplicar.
5. As a usuário, quero que cada grupo respeite a ordem de Rank do Vault, para a prioridade continuar visível.
6. As a usuário, quero que Issues `done` fiquem fora dos dois grupos, para a vista permanecer acionável.
7. As a usuário, quero que Issues blocked apareçam mesmo assim, para que atenção temporal não seja confundida com disponibilidade.
8. As a usuário, quero que `mt overdue` termine em silêncio com exit 0 quando nada se aplica, para o hábito diário e scripts continuarem simples.
9. As a usuário lembrado por `mt overdue`, quero que `mt undefer` sem argumentos limpe o `deferred_until` de todas as Deferrais expiradas do Vault, para arquivar todos os lembretes de uma vez.
10. As a usuário, quero que a limpeza em lote imprima `Undeferred <id> (was <datetime>)` por Issue, para saber o que foi arquivado.
11. As a usuário, quero que `mt undefer` sem expiradas termine em silêncio com exit 0, para rodar o hábito diário às cegas.
12. As a usuário que mudou de ideia, quero que `mt undefer <id>` limpe o `deferred_until` de uma Issue específica mesmo quando a Deferral ainda é futura.
13. As a usuário, quero que `mt undefer <id>` numa Issue sem `deferred_until` falhe com exit 1 e mensagem clara, para um alvo errado ser pego na hora.
14. As a usuário, quero que `mt undefer` toque apenas o campo `deferred_until` — Status e Rank intocados — para o comando nunca ter opinião sobre prioridade ou estado.
15. As a usuário, quero que `mt ready` e `mt pick-next` mantenham a semântica de disponibilidade atual, para o modelo do ADR-0002 ficar intacto.
16. As a usuário, quero que o glossário nomeie o conceito (Deferral expirada), para o vocabulário do projeto continuar preciso.

## Implementation Decisions

- **Conceito novo no glossário**: **Deferral expirada** — uma Deferred until cuja data/hora chegou (`now >= deferred_until`). `Avoid`: "vencida, acordou, lembrete-como-estado".
- **internal/list** (lógica pura) ganha:
  - `DeferralExpired(deferredUntil, now)` — valor parseável e não-futuro; espelho do `IsFutureDeferred` existente; vazio/malformado = false (validação de formato continua com `mt check`).
  - `ExpiredSuffix` → `[expirada MM-DD]` e `DeadlineSuffix` → `[deadline MM-DD]` (só data, família do `DeferSuffix` existente).
  - `OverdueGroups(items, now)` → duas partições ordenadas (expiradas, atrasadas): só não-`done`; Item com os dois sinais cai só no grupo das expiradas; a ordem de entrada (Rank) é preservada dentro de cada grupo.
- **internal/issue**: `Issue.Undefer()` limpa somente `DeferredUntil` (par do `Defer` existente).
- **CLI**:
  - `mt overdue` mantém o nome e ganha runner próprio: o runner compartilhado de query de predicado único não comporta saída em dois grupos; `mt ready` continua no runner compartilhado. A renderização reusa o formato de linha padrão, com os novos sufixos anexados.
  - Novo comando `mt undefer [id]`: sem argumentos → varre os Items do Vault limpando toda Deferred until expirada (reusando o predicado puro); com um ID → limpeza pontual pelo helper read-modify-write existente, com pré-checagem que falha (exit 1) quando o campo está vazio. Lote com zero correspondências imprime nada, exit 0. Linha de saída por Issue limpa: `Undeferred <id> (was <valor cru>)`.
- **Sem prompt de confirmação no lote**: um campo `deferred_until` expirado não tem mais efeito funcional (a disponibilidade já voltou), então limpá-lo não perde estado que importa.
- **Docs**: `CONTEXT.md` ganha o termo e estende a entrada de Deferred until; novo ADR-0006 registra a virada — `overdue` como atenção temporal e `undefer` como limpeza de campo, superando explicitamente a frase "não existe undefer" do ADR-0002 **sem** derrubar o modelo de disponibilidade dele.
- O sufixo `[defer ...]` de Deferrals futuras e as regras de ocultação de `mt list --all` ficam intocados.

## Testing Decisions

- Os dois seams estabelecidos do repo (TESTING.md) cobrem tudo; nenhum seam novo.
- **Seam 1 — processo CLI (primário)**: godog contra o binário compilado com Vault temporário, assertando stdout/stderr/exit code/arquivos. Atualizar os cenários de `overdue` em `ready-overdue.feature` (marcadores, agrupamento, dois-sinais-uma-vez, exclusão de `done`, inclusão de blocked, vazio em silêncio) e adicionar feature de `undefer`: lote limpa todas as expiradas e imprime os `was`; lote silencioso quando não há expiradas; per-id limpa Deferral futura; per-id sem campo sai com 1; Status/Rank intocados por `undefer`. Prior art: `ready-overdue.feature`, `issue.feature`.
- **Seam 2 — lógica pura exportada** (black-box, nunca internals): testes de unidade em `internal/list` para o predicado/sufixos/agrupamento novos (tabela, espelhando os testes de lista existentes) e teste de round-trip de `Undefer` em `internal/issue`. Ambos os pacotes já estão matriculados nos gates de coverage (≥90%) e mutation — sem mudança no Makefile.

## Out of Scope

- Lembretes com janela ("chegou hoje") — a persistência até `done` ou re-defer foi escolhida deliberadamente.
- `mt check` reportando Deferrais expiradas.
- Renomear o comando `overdue` ou criar aliases (memória muscular mantida).
- Saída JSON ou flags `--all` nas queries.
- Mudanças de cor/formato na renderização compartilhada de linha.
- Qualquer mudança na semântica de disponibilidade de `ready`/`pick-next`.

## Further Notes

- O design saiu de uma sessão de grilling (grill-with-docs): Q1–Q11 confirmadas pelo usuário; a feature é entregue em um único ticket de implementação.
- O `undefer` em lote é o passo "arquivar o lembrete" do fluxo diário: `mt overdue` → agir ou re-deferir → `mt undefer`.
- O valor impresso em `(was ...)` é o datetime naive cru armazenado (`YYYY-MM-DDTHH:MM`).
- Os dados migrados funcionam como estão: Deferrais já expiradas no Vault vivo (ex.: `dom-9wj`) aparecerão em `mt overdue` imediatamente após a mudança.
