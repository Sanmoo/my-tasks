Status: ready-for-agent

# Spec: `mt` — issue tracker pessoal (v1)

## Problem Statement

O usuário rastreia suas issues pessoais em arquivos Markdown (hoje com o nd + scripts Python próprios), mas o ferramental atual trava em vários pontos: o schema do nd é pesado para uso pessoal (`assignee`, `type`, `content_hash`, `History`, epics, dependências); adiar uma issue é um status em vez de um campo de data/hora (e a hora digitada é descartada); priorizar exige labels `rank/{area}/NNN` + campo `priority` e um script que dispara um subprocesso por mudança, levando segundos após fechar o editor; "áreas" são labels dentro de um único vault, quando na prática separam domínios que merecem ser vaults distintos. O usuário quer um tracker próprio, mínimo e git-friendly, com ordem de prioridade de primeira classe — e quer construí-lo como veículo para aprender a desenvolver com LLMs no dia a dia (ADR-0001).

## Solution

Um CLI em Go de binário único, `mt`, que opera sobre Vaults: diretórios com uma Issue por arquivo Markdown (`issues/<id>.md`, YAML frontmatter + corpo). Bookmarks globais endereçam vaults (`@nome`), com um favorito principal opcional; sem nenhum, o comando falha com instruções claras. Status mínimo (`open`/`in_progress`/`done`) extensível por vault; deferral é campo de data/hora — a Issue volta a ficar disponível sozinha; Rank é inteiro único por vault, e issues sem rank formam o Backlog. `prioritize` mantém o fluxo de $EDITOR com apply in-process instantâneo e reescrita só dos arquivos alterados; `pick-next` puxa o próximo disponível para `in_progress`. Comandos de apoio: `status`, `ready`, `overdue`, `check`, `comment`, `bookmark`. Tudo versionável em Git e legível por humanos e por LLMs.

## User Stories

1. As a usuário, quero inicializar um vault com `mt init`, para ter `issues/` e a configuração do vault prontos.
2. As a usuário, quero definir o prefixo de ID e a lista de status no init, para adaptar cada vault ao seu domínio.
3. As a usuário, quero criar uma issue com `mt create <título>`, para registrar uma unidade de trabalho em um arquivo Markdown.
4. As a usuário, quero que `created_at` seja carimbado automaticamente na criação, para nunca me preocupar com esse campo.
5. As a usuário, quero capturar rápido com `mt q <título>` (imprime só o ID), para registrar ideias sem sair do fluxo.
6. As a usuário, quero criar uma issue já com labels livres, para marcar assuntos transversais (ex.: `compras`).
7. As a usuário, quero que cada issue ganhe um ID curto e estável (prefixo do vault + sufixo aleatório), para referenciar issues sem colisão.
8. As a usuário, quero que criar (ou qualquer comando) sem vault definido falhe com mensagem clara (falta `@`, `--vault` ou favorito principal), para nunca escrever no lugar errado.
9. As a usuário, quero cadastrar bookmarks (`mt bookmark add <nome> <caminho>`), para endereçar vaults com `@nome`.
10. As a usuário, quero listar e remover bookmarks, para manter a configuração sob controle.
11. As a usuário, quero configurar um favorito principal, para os comandos funcionarem sem `@` no dia a dia.
12. As a usuário, quero que `@bookmark`, `--vault` e o favorito principal se resolvam nessa ordem de precedência, para saber exatamente qual vault será usado.
13. As a usuário, quero que a config global seja detectada automaticamente, para nunca passar caminhos no dia a dia.
14. As a usuário, quero definir status personalizados por vault, para modelar fluxos além de `open`/`in_progress`/`done`.
15. As a usuário, quero que `mt list` mostre as issues na ordem de prioridade (rank → backlog por `created_at`), para ver o que importa primeiro.
16. As a usuário, quero glyphs de status (○/◐/●) na listagem, para bater o olho e saber o estado de cada issue.
17. As a usuário, quero que `list` esconda `done` por padrão (com `--all` para ver tudo), para manter o foco no que está vivo.
18. As a usuário, quero que `list` mostre issues com Deferred until no futuro (sufixo `[defer ...]` sempre visível), para não esquecer do trabalho que ainda não está disponível.
19. As a usuário, quero filtrar por `--status` e `--label`, para recortar o vault.
20. As a usuário, quero ver uma issue completa (`mt show`), frontmatter + corpo + comentários, para ler todo o contexto de uma vez.
21. As a usuário, quero editar o arquivo via `mt edit` ($EDITOR), para ajustes livres preservando o resto do documento.
22. As a usuário, quero que uma edição manual não corrompa o frontmatter nas operações seguintes, para confiar que o formato é robusto.
23. As a usuário, quero fechar com `mt done` (carimbando `completed_at`), para registrar quando terminei.
24. As a usuário, quero usar `mt close` como alias de `done`, para digitar o que vier à cabeça.
25. As a usuário, quero reabrir com `mt reopen` (limpando `completed_at` e `started_at`), para corrigir fechamentos prematuros.
26. As a usuário, quero transitar livremente entre status com `mt status <id> <status>` (sem máquina de estados), para que só `pick-next`→`in_progress` e `done` terminal tenham comportamento especial.
27. As a usuário, quero adiar com `mt defer <id> <quando>` aceitando `YY-MM-DD HH:MM`, para tirar trabalho da fila acionável até o momento certo — **com hora**.
28. As a usuário, quero deferir com prazos relativos (`+2d`, `+1w`, `+3h`), para adiar sem calcular datas.
29. As a usuário, quero que a issue deferida continue `open` mas indisponível, para que deferral seja dado, não estado.
30. As a usuário, quero que a issue deferida volte a ficar disponível sozinha quando `now >= deferred_until`, para não existir operação de "undefer".
31. As a usuário, quero que `pick-next` ignore issues deferidas no futuro, para puxar apenas trabalho acionável.
32. As a usuário, quero priorizar com `mt prioritize` abrindo o $EDITOR com a lista `[P]`/`[ ]`, para manter meu fluxo atual de reordenação.
33. As a usuário, quero que reordenar as linhas `[P]` mude a ordem de prioridade, para priorizar apenas arrastando linhas.
34. As a usuário, quero promover `[ ]`→`[P]` e rebaixar `[P]`→`[ ]` no editor, para mover issues entre a fila e o Backlog sem outro comando.
35. As a usuário, quero que o editor inclua issues `open` e `in_progress`, para reordenar também o que está em andamento.
36. As a usuário, quero que os ranks sejam renormalizados 1..N ao salvar, para a fila nunca ter buracos.
37. As a usuário, quero que o apply do prioritize seja instantâneo (in-process, sem subprocesso por item), para não esperar segundos após fechar o editor.
38. As a usuário, quero que só os arquivos com rank alterado sejam reescritos, para o diff do Git conter só o que mudou de fato.
39. As a usuário, quero que uma edição inválida no editor (ID inexistente/duplicado) seja rejeitada sem aplicar nada, para nunca corromper a fila por engano.
40. As a usuário, quero `mt top <id>`, `mt bottom <id>` e `mt rank <id> <n>` para ajustes rápidos de um item, para priorizar sem abrir o editor.
41. As a usuário, quero `mt unrank <id>` para mandar uma issue de volta ao Backlog, para desafogar a fila.
42. As a usuário, quero que `pick-next` mova a issue `open` disponível de menor rank para `in_progress`, carimbando `started_at`, para começar a trabalhar com um comando.
43. As a usuário, quero que, sem issues rankeadas, `pick-next` pegue a mais antiga do Backlog, para nunca ficar sem próximo passo.
44. As a usuário, quero que `pick-next` sem nada disponível saia com mensagem clara e exit code 1, para scripts detectarem a situação.
45. As a usuário, quero que `pick-next` recuse com erro claro quando houver rank duplicado, para que ambiguidade de ordem nunca seja resolvida no chute.
46. As a usuário, quero poder ter várias issues `in_progress` ao mesmo tempo, para não inventar um limite de WIP que eu não pedi.
47. As a usuário, quero `mt ready` listando as disponíveis agora em ordem, para ver a fila acionável.
48. As a usuário, quero `mt overdue` listando issues com deadline estourado e não-`done`, para enxergar atrasos.
49. As a usuário, quero que `deadline` seja informativo (não bloqueia nada), para que prazos orientem sem engessar.
50. As a usuário, quero `mt check` reportando rank duplicado como erro e lacunas como aviso brando, para saber o estado de integridade do vault.
51. As a usuário, quero `mt check` validando YAML malformado, status fora da config e formato de datetime, para pegar erros de edição manual.
52. As a usuário, quero `mt check --fix` renormalizando ranks, para curar o vault com um comando.
53. As a usuário, quero `mt comment <id> [texto]` adicionando comentário com timestamp, para registrar progresso sem editar o arquivo na mão.
54. As a usuário, quero que cada comentário tenha uma âncora estável, para editar/apagar comentários depois com diffs limpos.
55. As a usuário, quero que comentários sejam append-only sobre o corpo existente, para nunca reescrever conteúdo que não mudou.
56. As a usuário, quero instalar o `mt` como binário único sem runtime externo, para usar em qualquer máquina.
57. As a usuário, quero mensagens de erro no stderr com exit codes distintos de sucesso, para compor comandos em scripts.
58. As a usuário, quero que todas as datas sejam gravadas como `YYYY-MM-DDTHH:MM` naive (sem timezone/segundos), para diffs mínimos e leitura humana.
59. As a usuário, quero que não exista `updated_at` armazenado (o Git/o mtime dizem isso), para que cada edição mexa só no que mudou.
60. As a usuário, quero que o campo `id` não exista no frontmatter (o nome do arquivo é a autoridade), para nunca dessincronizar id de nome de arquivo.

## Implementation Decisions

- **Linguagem/entrega:** Go, binário único, sem runtime externo (ADR-0001). CLI com padrão de aliases (`done`/`close`).
- **Modelo de domínio:** segue o vocabulário do `CONTEXT.md` — Issue, Vault, Bookmark, Rank, Backlog, Deferred until, Deadline, Status. Uso de sinônimos evitados (ex.: "task", "área") é proibido em código, mensagens e testes.
- **Schema da Issue** (frontmatter YAML + corpo Markdown; forma exata como decidida na grelha):

```yaml
---
title: "[Niver Edu] comprar material de Assaí conforme doc compartilhado"
status: open
labels: [compras, familia]
created_at: 2026-08-15T09:30
rank: 2
deferred_until: 2026-08-20T08:00
deadline: 2026-08-22T18:00
---

## Description
## Notes
## Comments
### 2026-08-16T14:05
Comprei metade da lista.
<!-- comment: 4f2b9c1a -->
```

- **Regras de campo:** sempre presentes: `title`, `status`, `labels`, `created_at`; presentes só quando têm valor: `rank`, `deferred_until`, `deadline`, `started_at`, `completed_at`. Sem `id` (nome do arquivo é a autoridade). Sem `updated_at` (Git/mtime). Datetime `YYYY-MM-DDTHH:MM` naive. Seções do corpo: apenas `Description`, `Notes`, `Comments`. Comentário = heading de timestamp + âncora estável `<!-- comment: <curto> -->`.
- **IDs:** prefixo do vault + sufixo curto aleatório (ex.: `pkm-055`); colisão verificada na criação.
- **Config global** (auto-detectada, caminho XDG `~/.config/mt/`), com shape:

```yaml
default: bjd
bookmarks:
  bjd: ~/dev/github.com/Sanmoo/pkm/.vault
  dom: ~/dev/github.com/Sanmoo/dom/.vault
```

- **Config do vault** (`mt.yaml` na raiz), com shape:

```yaml
prefix: PKM
status: [open, in_progress, done]
```

- **Resolução de vault:** `@bookmark` > `--vault <path>` > `default`. Sem nenhum: erro com instruções (falta `@`, `--vault` ou favorito principal). **Sem detecção por cwd** (ADR-0003).
- **Status:** `open`/`in_progress`/`done` + customizados via config do vault. Sem FSM: transições livres via `mt status <id> <status>` (valida contra a lista configurada do vault); apenas `pick-next` (→ `in_progress`) e `done` (terminal, carimba `completed_at`) têm comportamento especial. `reopen` limpa `completed_at` e `started_at`. Status fora da config é reportado por `check`.
- **Deferral:** campo, não status (ADR-0002). Disponibilidade = `now >= deferred_until`; sem operação de undefer. `mt defer` aceita `YY-MM-DD HH:MM` e relativos (`+2d`, `+1w`, `+3h`).
- **Rank:** inteiro único por vault; ausente = Backlog. Duplicado = **erro de usuário**: `check` reporta, `list` mostra warning, `pick-next` recusa; lacunas = aviso brando. Cura = renormalização 1..N via `prioritize` (ao salvar) ou `check --fix`. Renormalização reescreve **apenas** os arquivos cujo rank mudou (zero churn nos intactos).
- **`prioritize`:** fluxo $EDITOR com buffer neste formato (formato de decisão, herdado do fluxo nd-prioritize do usuário):

```text
# Edit ranking for this vault
# Reorder lines. Use [P] for prioritized, [ ] for backlog.
# Do not edit issue IDs. Save and close to continue.

[P] pkm-055  <título>
[P] pkm-07r0  <título>
[ ] pkm-5qa8  <título>
```

Inclui issues `open` e `in_progress`. Reordenar `[P]` muda a ordem; trocar `[ ]`↔`[P]` entra/sai da fila. Conteúdo inválido (ID inexistente/duplicado) → erro, nada é aplicado. Apply **in-process** (requisito de performance — a dor original era N subprocessos `nd update`).

- **`pick-next`:** candidatos = `open` + disponível; ordem: menor rank → mais antigo do Backlog por `created_at` → `id`. Seta `in_progress` + `started_at`. Sem candidatos: stderr + exit 1. Rank duplicado: recusa, exit 1. Sem WIP limit.
- **`list`:** glyphs ○ (`open`), ◐ (`in_progress`), ● (`done`), fallback para custom; ordem rank → Backlog por `created_at`; esconde só `done` (sufixo `[defer MM-DD HH:MM]` sempre visível nas deferidas-futuras); `--all` inclui `done`; `--status`, `--label`.
- **Exit codes:** 0 sucesso; 1 erro de usuário (vault indefinido, nada disponível, rank duplicado, edição inválida); 2 erro de uso. Erros sempre em stderr.
- **`check`:** valida rank duplicado (erro), lacuna (aviso), YAML/frontmatter malformado, status fora da config, formato de datetime; `--fix` renormaliza.

## Testing Decisions

- **Filosofia:** testar **comportamento externo**, nunca detalhes de implementação — nada de testes que alcançam símbolos não-exportados ou assertam estrutura interna.
- **Seam 1 (primário, o mais alto possível) — o processo CLI:** e2e com Gherkin via Godog, executando o **binário compilado** contra um vault temporário por cenário; `$EDITOR` fake (script auxiliar que escreve um buffer preparado) para testar `prioritize`/`edit` headless. Asserções: stdout, stderr, exit code e o estado dos arquivos em disco. Cada user story vira um cenário Gherkin.
- **Seam 2 (justificado, ponto mais alto abaixo do CLI) — APIs exportadas da lógica pura:** testes unitários black-box das packages densas em decisão: parse de datetime (`YY-MM-DD HH:MM`, relativos), round-trip do frontmatter (campos opcionais presentes/somente-quando-setados, ordem estável), renormalização de ranks (1..N + diff mínimo de reescrita), precedência de resolução de vault (`@` > `--vault` > `default` > erro), geração de IDs. Apenas símbolos exportados.
- **Cobertura:** gate de alta cobertura (≥90%) **nas packages do Seam 2**, onde a métrica mede lógica real. Sem gate global de % — o wiring de comandos (Cobra) é coberto por comportamento no e2e, não por coverage de unit.
- **Mutation testing:** gremlins rodando contra os testes unitários do Seam 2, como gate de qualidade desses testes. Mutação do glue de comandos está fora do escopo de mutação (ali quem mata mutante é o e2e).
- **Prior art:** não há testes neste repo (vazio). Referência mais próxima: o harness de teste do usuário em `pkm/scripts` (Python, subprocesso) — o Seam 1 é a evolução Go desse padrão, com Gherkin em vez de asserts manuais.
- **Ferramentas:** `testing` stdlib + `godog` (Gherkin) + `gremlins` (mutação) + `go test -cover` (gate do Seam 2).

## Out of Scope

Dependências computadas (ADR-0004): `blocked_by` é campo e o estado blocked é derivado. Fora do escopo: epics, full-text search, TUI de ordenação, migração do vault nd (`import-nd` — one-off pós-validação do formato), multi-usuário/assignee, sync multi-branch, armazenar `updated_at`, `content_hash`, export/import JSON, plugins.

## Further Notes

- Docs de domínio já existentes: `CONTEXT.md` e ADRs 0001–0003 em `docs/adr/`; esta spec usa o vocabulário do glossário e respeita os ADRs.
- O repo ainda não é um repositório Git — inicializar (`git init`) antes da implementação; o design inteiro assume git-friendliness.
- Projeto é também veículo para aprender desenvolvimento assistido por LLMs (ADR-0001) — decisões de processo de desenvolvimento podem virar ADRs futuros.
- A spec é o desdobramento do design consolidado via `/grill-with-docs`; em conflito entre esta spec e o design da grelha, vale a spec e o conflito deve ser reportado.
