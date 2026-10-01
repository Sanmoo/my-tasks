# Comentário de transição é Comment no corpo, não campo

O nd guardava `close_reason` no frontmatter e o ADR-0005 registrou que o mt
**descartou** esse campo na migração. As transições de status passaram a
aceitar um comentário final opcional (`done`/`close`, `reopen`, `status`), e
a pergunta que o descarte deixa aberta é onde esse texto vive: um campo de
frontmatter (o `close_reason` de volta, agora sob outro nome) ou um Comment
no corpo. Escolhemos o **Comment no corpo**, na mesma gravação da transição:
é o artefato que `mt comment` já escreve — heading com timestamp, texto,
âncora estável —, então não há segundo formato a renderizar, validar ou
migrar, e o timestamp do heading coincide com o `completed_at` da transição
que o carrega. O preço aceito é que o texto não é consultável como campo: um
`close_reason` apareceria em `show`/`list`/jq sem parsing, enquanto o
comentário só é lido pelo que já lê Comments (`mt show --summary` mostra o
último como `Last comment`). Como comentários são append-only, o texto
corrigido depois vira uma correção registrada, não uma reescrita — que é o
comportamento certo para o registro de por que algo mudou de estado. Um
comentário presente mas em branco é recusado (exit 2) justamente porque a
gravação é irreversível.
