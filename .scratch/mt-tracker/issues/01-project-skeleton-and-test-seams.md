# 01 — Esqueleto do projeto e seams de teste

**What to build:** módulo Go com o binário `mt` compilável e toda a infraestrutura de teste da spec: e2e com Gherkin (Godog) executando o **binário compilado** contra vaults temporários, com `$EDITOR` fake para fluxos de editor headless; testes unitários black-box das APIs exportadas da lógica pura; alvo único que roda unit + e2e + coverage + mutação (gremlins); gate de cobertura ≥90% nas packages de lógica pura (sem gate global de %).

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [x] `mt --help` imprime ajuda e sai com exit 0
- [x] Cenário Gherkin de smoke roda verde via binário compilado (não via função chamada)
- [x] Helper de vault temporário e `$EDITOR` fake disponíveis para todos os cenários
- [x] Um alvo de automação roda unit + e2e + coverage + mutação
- [x] Gate de cobertura ≥90% configurado sobre as packages de lógica pura
- [x] Convenção de exit codes (0 sucesso / 1 erro de usuário / 2 erro de uso) estabelecida e respeitada no smoke

## Comments

### Implementação

- Módulo `github.com/Sanmoo/my-tasks2`, Go 1.26.3; binário em `cmd/mt`, wiring cobra em `internal/cli` (sem gate de coverage — coberto por e2e).
- `internal/exitcode`: package de lógica pura com a convenção 0/1/2 + mapeamento de erros tipados (`Usage`/`User`/`For`), 100% de coverage unitário black-box e 100% de eficácia no gremlins.
- e2e (`e2e/`): TestMain compila o binário uma vez e roda o godog; `features/smoke.feature` tem 5 cenários (help/exit 0, unknown command e unknown flag → exit 2, vault temporário e fake editor disponíveis em todo cenário); `support/` tem os helpers de vault, `$EDITOR` fake (printf, conteúdo byte a byte) e isolamento de `XDG_CONFIG_HOME`.
- `make check` = unit + e2e + coverage-gate (≥90% por package de lógica pura, sem gate global) + gremlins. Thresholds de mutação em `.gremlins.yaml` — os flags float64 do gremlins v0.6.0 estão quebrados (viper lê como string, nunca gata).
- Notas de spec: (1) exit 1 não tem caminho exercitável no smoke ainda — nenhum comando do esqueleto produz erro de usuário; a convenção está estabelecida e testada no unit. (2) A lista de Seam 2 da spec não cita exit codes; `internal/exitcode` entrou na lista de packages puras porque o ticket pede a convenção estabelecida — delta registrado no TESTING.md.
- `TESTING.md` documenta seams, targets e contratos do harness (inclusive a limitação de que o step `I run` ainda não tem tokenização com aspas).
