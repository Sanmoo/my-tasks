# 17 — show com saída colorida (estilo nd show)

**What to build:** `mt show <id>` passa a renderizar a Issue em uma vista estruturada e colorida, no estilo do `nd show`: linha de cabeçalho com glyph de status, ID, título em negrito e status colorido; linhas de metadados com rótulos em cor de destaque (Created, Labels e os opcionais só quando set: Rank, Deferred until, Deadline, Started, Completed, Blocked by); corpo Markdown renderizado com glamour (tema dark/light, wrap ≤100). Cor respeita a convenção padrão: `NO_COLOR` desliga, `CLICOLOR=0` desliga, `CLICOLOR_FORCE` força, senão só quando stdout é TTY. Sem cor (pipe), a vista é a mesma porém sem ANSI — nunca mais o dump cru do arquivo (igual ao `nd show`).

**Motivo:** o `mt show` atual despeja o arquivo cru; o usuário quer leitura colorida como no `nd show`. (ADR-0001: construímos o nosso, mas o formato de saída do nd é o modelo visual.)

**Fora de escopo:** `--short`/`--json` (o nd tem; não pedidos), colorir `mt list`, raw dump via flag.

**Blocked by:** —

**Status:** resolved

- [x] Pacote pure-logic `internal/show` com `Render(issue, id, color)` — header, metadados (opcionais só quando set), corpo (glamour quando color, cru quando não)
- [x] `mt show` parseia e renderiza; detecção de cor (NO_COLOR / CLICOLOR / CLICOLOR_FORCE / TTY) no cli
- [x] Unit tests black-box de `internal/show` (glyphs, cores por status, campos opcionais, ANSI on/off)
- [x] Cenários e2e de `show` atualizados para a vista estruturada (sem cor: header, metadados, corpo cru)
- [x] `internal/show` em PURE_PACKAGES (coverage ≥90% + mutation)

### Implementado

`internal/show` (novo pacote pure-logic, 96.6% coverage, 97.3% efficacy no gremlins — o único mutante vivo é equivalente: `> 100` vs `>= 100` no cap de wrap). `Render(issue, id, Options{Color, Width})`: header `◐ id . título [status]` com glyph/cores da paleta ayu do nd, metadados (Created, Labels, Rank, Deferred until, Deadline, Started, Completed, Blocked by — opcionais só quando set, datetimes T→espaço), corpo via glamour (dark/light por `MT_THEME`/`COLORFGBG`/default dark, wrap min(largura, 100), default 80). `ShouldUseColor(tty)` segue no-color.org: `NO_COLOR` > `CLICOLOR=0` > `CLICOLOR_FORCE` > TTY. No cli, `show` parseia (issue malformada agora é erro exit 1, não dump cru) e detecta TTY/largura via `x/term`; a vista estruturada vale também em pipe, sem ANSI (como o `nd show`). E2e: cenários atualizados (vista estruturada, campos opcionais, ANSI via `CLICOLOR_FORCE`); harness ganhou step `the environment variable ... is ...` e expansão de placeholders em `stdout matches`/`stdout does not contain` (bug latente — `does not contain "<id>"` nunca expandia). Deps novas: `charmbracelet/glamour` + `golang.org/x/term`. README atualizado.
