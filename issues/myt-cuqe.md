---
title: 'Suportar claimed_by: claim como campo, não como status'
status: open
labels: [ready-for-agent]
created_at: 2026-10-15T10:13
rank: 1
---

## Description

### Problem Statement

Eu uso as skills de engenharia (triage, to-tickets, wayfinder) sobre o tracker de issues deste projeto, e elas precisam de uma noção de **claim**: assumir um ticket **antes** de começar o trabalho, para que sessões concorrentes não peguem o mesmo item. Hoje o `mt` não tem como registrar isso. O único estado parecido é `in_progress` — mas `in_progress` significa **"o trabalho começou"** (é o que `pick-next` grava), não "alguém assumiu e ainda não começou". Sobrepôr os dois significados deixa o estado "assumida mas não iniciada" irrepresentável, e sessões paralelas disputam issues sem um gate real no nível da ferramenta.

### Solution

Um campo de frontmatter opcional `claimed_by` (string) no Issue, com dois comandos novos — `mt claim <id> <who>` e `mt unclaim <id>` — e uma extensão do predicado de disponibilidade: uma Issue com claim ativo **não é Disponível**, então `ready` e `pick-next` a pulam. O claim é um **campo, não um Status** (mesma família conceitual de `deferred_until` e `blocked_by`), e a posse é de sessão, não de pessoa: o vault é pessoal; o valor do campo distingue *que* sessão claimou, não *quem* (não há times no domínio).

### Job Stories

1. Quando duas sessões de agente trabalham sobre o mesmo vault e uma quer reservar um ticket antes de iniciar, eu quero executar `mt claim <id> <who>`, para que a outra sessão não escolha o mesmo ticket.
2. Quando uma sessão está escolhendo o próximo trabalho com `pick-next`, eu quero que Issues com claim ativo sejam puladas automaticamente, para que a sessão nunca inicie um ticket já reservado por outra.
3. Quando rodo `mt ready` para ver o que posso pegar agora, eu quero que tickets claimados não apareçam, para que a lista reflita só o que de fato está livre.
4. Quando uma sessão claimou por engano ou o trabalho foi abortado, eu quero `mt unclaim <id>` para devolver o ticket ao pool de disponíveis.
5. Quando uma sessão claima um ticket e depois o inicia com `pick-next`, eu quero que `in_progress` e `claimed_by` coexistam sem se confundirem, para que o histórico mostre quem assumiu e quando começou.
6. Quando termino ou reabro um ticket, eu quero que o ciclo do campo seja previsível: `done` preserva o claim como histórico e `reopen` o limpa, para que um ticket reaberto volte realmente ao estado "ninguém pegou".
7. Quando um claim fica órfão numa issue `done` e quero corrigi-lo manualmente, eu quero poder `mt unclaim` sobre qualquer estado, para que correções raras não exijam hacks de arquivo.
8. Quando executo `mt claim` sem informar quem claimou, eu quero um erro de uso (exit 2), para que o campo nunca perca a informação de posse — que é exatamente a sua razão de existir.
9. Quando tento claimar uma Issue que não está `open` (in_progress, done, status customizado), eu quero um erro de usuário (exit 1), para que claim continue significando "reserva de trabalho ainda não iniciado".
10. Quando tento `mt unclaim` numa Issue sem claim, eu quero um erro de usuário (exit 1), para que operações sobre estado inexistente sejam falhas explícitas e não silênciosos.
11. Quando listo as issues com `mt list`, eu quero ver um sufixo `[claimed]` nas issues com claim, para que a fila mostre a posse de relance sem virar uma parede de texto.
12. Quando examino uma issue com `mt show`, eu quero ver o valor completo de `claimed_by`, para que eu saiba qual sessão tem o ticket.
13. Quando abro o arquivo da Issue para ler o frontmatter, eu quero que `claimed_by` esteja lá como campo opcional, para que o vault continue legível e versionável em Git sem ferramenta.
14. Quando rodo `mt check` sobre o vault, eu quero que o campo seja validado (string vazia ou só espaços é inválida), para que corrupção manual de frontmatter seja detectada no audit.
15. Quando escrevo a configuração do tracker (`docs/agents/issue-tracker.md`) e as skills mandam "assign the ticket to the driving dev", eu quero uma tradução documentada (`mt claim <id> <who>`), para que nenhuma skill improvise uma convenção própria.
16. Quando o tracker passa a ser o próprio `mt` neste repo, eu quero o vault `myt` bootstrapado dentro do repo com o primeiro ticket, para que o dogfooding comece pela issue do próprio tracker.
17. Quando decidir depois se o modelo se generaliza, eu quero um ADR registrando por que claim é campo e não status e por que o campo é `claimed_by` e não `assignee`, para que leitores futuros entendam o trade-off sem me perguntar.

### Implementation Decisions

- **Claim é campo, não Status**: campo opcional `claimed_by` (string) no schema de frontmatter do Issue. Não há máquina de estados; sem transições. Segue a família estabelecida pelos ADRs existentes ("defer é campo, não status"; "dependências são computadas").
- **`mt claim <id> <who>`**: grava `claimed_by = <who>`. `<who>` é obrigatório — ausente é usage error (exit 2). Só permitido em Issue `open`; qualquer outro status é user error (exit 1). Issue já claimada por outro valor: o novo claim sobrepõe (a posse pode ser transferida manualmente), sem erro.
- **`mt unclaim <id>`**: limpa o campo. Em Issue sem claim é user error (exit 1). Permitido em qualquer status (correção manual), inclusive `done`.
- **Disponibilidade**: o predicado canônico passa a incluir "sem claim ativo". `ready` e `pick-next` pulam Issues com `claimed_by` preenchido, exatamente como pulam blocked. Isso é mudança de comportamento nas duas superfícies.
- **Ciclo de vida**: `reopen` limpa `claimed_by` junto de `completed_at`/`started_at`; `done` preserva. `status` e as demais transições não tocam no campo.
- **Renderização**: `mt list` adiciona sufixo `[claimed]` nas issues claimadas (como o sufixo `[blocked]`); `mt show` exibe o valor completo. A linha de lista continua curta — o valor só aparece no `show` e no arquivo.
- **Validação**: `mt check` valida o campo — presente mas vazio/só espaços é erro de audit.
- **Sem comentário anexado a claim/unclaim**: claim não é transição de status (ADR-0007 trata do comment em transição); os comandos não aceitam trailing comment.
- **`pick-next` não auto-claima**: `pick-next` continua gravando apenas `in_progress`; o claim é um ato explícito da sessão.
- **Bootstrap do vault**: vault `myt` dentro deste repo (`issues/` + `mt.yaml`), status padrão `[open, in_progress, done]`. O ticket desta spec é a primeira Issue do vault.
- **Documentação**: novo termo **Claim** no glossário e ajuste do termo **Disponível**; ADR-0010 (claim como campo, não status; `claimed_by` vs `assignee`); reescrita de `docs/agents/issue-tracker.md` para o `mt` como tracker canônico de issues e specs (GitHub permanece apenas para PRs, flag "PRs as request surface" continua `no`); atualização do bloco `## Agent skills` no `AGENTS.md`. `docs/agents/triage-labels.md` não muda: os cinco rótulos canônicos já existem como labels free-form do `mt`.
- **Convenções documentadas no issue-tracker.md**: mutação de labels pós-criação por edição direta do frontmatter seguida de `mt check`; wayfinder: label `wayfinder:map` no issue-mapa, linha `Part of <map-id>` no body do filho, blocking via `mt dep`, frontier via `mt ready` ∩ children.

### Testing Decisions

- Bom teste aqui exercita só comportamento externo: stdout/stderr e exit codes do binário `mt`, mais o estado do vault em disco (frontmatter antes/depois) — nunca detalhes internos de implementação.
- Seam única: o binário `mt` via harness Gherkin e2e existente, em feature própria (`claim.feature`) com prior art em `defer.feature` (campo temporal no schema e no ciclo de vida), `dependency.feature` (gate computado de disponibilidade em ready/pick-next/list) e `status.feature` (erros de estado e exit codes).
- Cenários e2e a cobrir: claim grava o campo; claim pulado por ready/pick-next; unclaim devolve ao pool; claim em não-open rejeitado (exit 1); claim sem `<who>` rejeitado (exit 2); unclaim sem claim rejeitado (exit 1); reopen limpa; done preserva; sufixo `[claimed]` no list; validação pelo `mt check`; frontmatter legível em disco.
- Testes unitários acompanham o padrão da casa: schema do Issue (round-trip do campo), validação do check, e o predicado de disponibilidade no pacote onde ele vive hoje.

### Out of Scope

- Qualquer noção multi-usuário: times, permissões, `assignee` como identidade de pessoa.
- `pick-next` gravando claim automaticamente.
- Sub-issues nativas / hierarquia pai-filho no schema (wayfinder continua por convenção de body + label).
- Comando para mutação de labels (`mt label add/rm`) — a convenção documentada é edição direta do frontmatter.
- Suporte a `--label` em `mt q` (o `mt create` cobre).
- Migração de tickets antigos de `.scratch/` ou GitHub para o vault `myt`.

### Further Notes

- O default bookmark do usuário aponta para outro vault (`bjd`, no pkm); todas as operações no vault `myt` devem usar `--vault` explícito ou bookmark criado pós-merge — resolução por cwd não existe.
- O valor do claim é texto livre: nesta primeira iteração, sessões de agente usam o próprio identificador de sessão.
- A triagem de rótulos (`ready-for-agent`) já está aplicada a este ticket — spec completa, pronta para implementação AFK.

## Notes
## Comments
### 2026-10-15T11:05
Implementado (commits 1492406..b4cbdc8): campo claimed_by, mt claim/unclaim, Disponibilidade com claim, [claimed] no list, Claimed by no show, check de claim em branco, ADR-0010, e2e claim.feature (14 cenários), README. Reviewer dual (standards+spec) sem blockers. PR #17 aguardando merge.
<!-- comment: f3762605 -->
