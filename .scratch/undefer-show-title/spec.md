Status: ready-for-agent

# Spec: `mt undefer` imprime o título da Issue

## Problem Statement

O `mt undefer` imprime `Undeferred <id> (was <datetime>)` por Issue. Quem roda o lote (o passo "arquivar o lembrete" do fluxo diário `mt overdue` → agir → `mt undefer`) vê só o ID: para saber o que foi arquivado precisa de um `mt show` por ID. As transições (`done`, `reopen`, `status`, `pick-next`) já imprimem o título na linha de confirmação; o `undefer` é a exceção que só mostra o ID.

## Solution

A linha do `mt undefer` (lote e per-id) passa a terminar com o título da Issue no formato da casa: `Undeferred <id> (was <datetime>): <título>`. O bloco `(was ...)` fica intocado, o título entra como sufixo `: <título>` (mesmo truque das transições), whitespace do título colapsado em espaço único, e título vazio mantém a linha de hoje — sem `: ` pendurado. Nada mais muda: exit codes e semântica de Status/Rank intocadas.

## User Stories

1. As a usuário que roda `mt undefer` em lote, quero que cada linha `Undeferred` termine com `: <título>` da Issue, para saber o que foi arquivado sem abrir cada Issue.
2. As a usuário, quero que o título venha **depois** do bloco `(was <datetime>)`, para o payload existente permanecer no formato que scripts já consomem.
3. As a usuário, quero que o mesmo formato valha no modo per-id (`mt undefer <id>`), para a confirmação pontual também identificar a Issue.
4. As a usuário com título contendo quebras de linha ou espaços corridos, quero que o título impresso tenha o whitespace colapsado em espaço único, para a linha nunca quebrar em múltiplas linhas.
5. As a usuário com Issue sem título (ou só whitespace), quero que a linha fique exatamente como hoje (`Undeferred <id> (was ...)`), para não aparecer um `: ` pendurado.
6. As a usuário, quero que nada mais mude no comando — exit codes, silêncio no lote vazio, falha do per-id sem campo — para o hábito diário continuar intacto.
7. As a usuário, quero que `mt overdue`, `mt ready` e `mt pick-next` permaneçam como estão, para a mudança ficar restrita ao `undefer`.

## Implementation Decisions

- A linha do `undefer` ganha o sufixo `: <título>` no fim, espelhando o sufixo das transições (`<id> is now <status>: <título>`): colapso de whitespace (\r\n, \r, \n → espaço único) + TrimSpace; título vazio → sem sufixo.
- O bloco `(was <datetime cru>)` permanece no fim da linha, antes do sufixo — posição e valor inalterados.
- A renderização compartilha a lógica de colapso/sufixo com a linha de transição (um único helper para "título como sufixo"), em vez de uma segunda cópia do truque.
- Os dois modos (lote e per-id) usam a mesma renderização de linha.
- Nenhuma mudança em `internal/list`, `internal/issue` ou nos predicados de disponibilidade — mudança é só de renderização no processo CLI.
- Docs: README (seção do `mt undefer`, exemplo de saída) passa a mostrar o formato com título. O spec do `overdue-undefer` e o ADR-0006 ficam como estão: registram a decisão de design da virada, e o formato exato de linha é detalhe de renderização.

## Testing Decisions

Seams: os dois seams estabelecidos (TESTING.md), nenhum seam novo.

- **Seam 1 — processo CLI (primário)**: godog contra o binário compilado. Atualizar `undefer.feature`: os asserts de linha do lote e do per-id passam a incluir o título; cenários novos para título com whitespace (colapso em linha única) e título vazio (linha sem sufixo). Prior art: os cenários existentes de `undefer.feature` e os de transições com título.
- **Seam 2 — lógica pura exportada**: sem mudança — não há lógica pura nova; o colapso de título permanece no processo CLI, coberto pela Seam 1 (mesmo arranjo da linha de transição). Nenhum pacote novo, nenhuma mudança no Makefile/gates.
- As histórias 1–5 viram asserts de stdout; 6–7 viram asserts de não-regressão nos cenários existentes.

## Out of Scope

- Mostrar título em `mt overdue` ou em qualquer outra saída.
- Mudar o formato do `(was ...)` ou adicionar flags de formato/JSON.
- Reescrever o spec do `overdue-undefer` ou o ADR-0006.
- Alterar a linha das transições além do necessário para compartilhar o helper.

## Further Notes

- A decisão saiu de sessão de grilling (Q1–Q4 confirmadas: formato com título no fim via `: `, escopo nos dois modos, título vazio sem sufixo, ticket novo sem emendar ADR/spec antigo).
- O `(was ...)` continua imprimindo o datetime naive cru armazenado (`YYYY-MM-DDTHH:MM`).
