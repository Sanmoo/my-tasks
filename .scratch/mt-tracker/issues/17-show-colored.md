# 17 — show com saída colorida (estilo nd show)

**What to build:** `mt show <id>` passa a renderizar a Issue em uma vista estruturada e colorida, no estilo do `nd show`: linha de cabeçalho com glyph de status, ID, título em negrito e status colorido; linhas de metadados com rótulos em cor de destaque (Created, Labels e os opcionais só quando set: Rank, Deferred until, Deadline, Started, Completed, Blocked by); corpo Markdown renderizado com glamour (tema dark/light, wrap ≤100). Cor respeita a convenção padrão: `NO_COLOR` desliga, `CLICOLOR=0` desliga, `CLICOLOR_FORCE` força, senão só quando stdout é TTY. Sem cor (pipe), a vista é a mesma porém sem ANSI — nunca mais o dump cru do arquivo (igual ao `nd show`).

**Motivo:** o `mt show` atual despeja o arquivo cru; o usuário quer leitura colorida como no `nd show`. (ADR-0001: construímos o nosso, mas o formato de saída do nd é o modelo visual.)

**Fora de escopo:** `--short`/`--json` (o nd tem; não pedidos), colorir `mt list`, raw dump via flag.

**Blocked by:** —

**Status:** ready-for-agent

- [ ] Pacote pure-logic `internal/show` com `Render(issue, id, color)` — header, metadados (opcionais só quando set), corpo (glamour quando color, cru quando não)
- [ ] `mt show` parseia e renderiza; detecção de cor (NO_COLOR / CLICOLOR / CLICOLOR_FORCE / TTY) no cli
- [ ] Unit tests black-box de `internal/show` (glyphs, cores por status, campos opcionais, ANSI on/off)
- [ ] Cenários e2e de `show` atualizados para a vista estruturada (sem cor: header, metadados, corpo cru)
- [ ] `internal/show` em PURE_PACKAGES (coverage ≥90% + mutation)
