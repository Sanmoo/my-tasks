# A priorização consome a linha única de Issue — o standalone quebrado de propósito

`internal/priority` nasceu **standalone**: não importava nada do
repositório, nem a definição de Issue. Para ordenar, mantinha uma
**projeção paralela** — um tipo próprio com ID, título, status, Rank e
`created_at` — e uma cópia própria da regra de Fila: menor Rank primeiro,
Backlog depois por `created_at`, ID como desempate final. O mesmo algoritmo
vivia também em `internal/list` (o `Sort` das vistas): duas
implementações do mesmo invariante, e a verdade da ordem não tinha casa.

A pergunta previsível é "por que quebrou o standalone?". Quebrou porque o
standalone **não era um contrato, era um acidente de história**: o pacote
nasceu antes de a Issue virar linha única e nunca precisou falar sobre ela,
só *parecer* com ela. Manter isso tem um preço concreto — a projeção
precisa ser preenchida à mão em cada porta de entrada (todo comando que
quisesse priorizar escrevia um adaptador de Issue para a projeção) e a
regra de Fila precisa ser mantida em duas implementações que podem
divergir em silêncio. Em troca, o standalone não comprava nada: nenhuma
view depende da priorização, nenhum pacote é reutilizado por outro
projeto, e o pacote já morava dentro do mesmo módulo.

O trade-off real era este: **duplicar a regra de Fila em cinco usos
internos** — o Buffer do `mt prioritize`, o Encaixe do `mt place`, a sessão
implícita do `mt create`/`mt q`, os planos rápidos do `mt top`/`mt
bottom`/`mt rank`/`mt unrank` e a renormalização do `mt check --fix` —
mais uma projeção mantida à mão para cada um, **versus uma borda nova,
única e descendente** sobre a base do domínio. A borda nova foi escolhida:
uma dependência só, na direção que o domínio já impõe (`internal/list`
depende de `internal/issue` pelo mesmo motivo), contra cinco acoplamentos
ao mesmo invariante que a borda elimina.

A quebra é **para baixo, não para o lado**. `internal/priority` importa
`internal/issue` e nada mais do repositório: consome `issue.Item` (a linha
única com ID + Issue), `issue.Compare` (a regra de Fila) e
`issue.RenormalizeRanks` (a renormalização 1..N), e o plano que devolve é
o `issue.Change` do domínio. A projeção some: `Buffer`, `Plan`,
`QuickPlan`, `Candidates`, `PlacementTarget`, `NewPlace` e `Place` passam a
receber e devolver `issue.Item`, lendo `Title`, `Status`, `Rank` e
`CreatedAt` do frontmatter. No CLI, os adaptadores que enfiavam Issues na
projeção (`priorityIssueFrom`, `priorityIssuesFromItems`,
`priorityIssuesFromCheckItems`, `loadPriorityIssues`) **deixam de existir**,
e o `mt check --fix` alcança `issue.RenormalizeRanks(items)` direto. Nada
mais depende dele, nada é invertido, e a ordem passa a ter uma casa só —
`internal/issue` — testada lá.

Os mecanismos internos ficam onde estão e com a função que tinham: o Buffer
e seu Parse, as Comparações do Encaixe, os planos rápidos. É uma troca de
casa, não uma reescrita — o `mt prioritize`, o `mt place` e o `mt check
--fix` se comportam exatamente como antes, e os ornamentos (`[defer]`,
`[blocked]`, `[expirada]`, `[deadline]`) ficam onde estavam. Consequência:
a ordem é uma decisão do domínio, e quem a consome apenas decide *quando*
perguntar e *quais* posições mover.
