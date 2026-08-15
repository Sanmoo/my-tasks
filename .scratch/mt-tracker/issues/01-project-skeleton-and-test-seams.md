# 01 — Esqueleto do projeto e seams de teste

**What to build:** módulo Go com o binário `mt` compilável e toda a infraestrutura de teste da spec: e2e com Gherkin (Godog) executando o **binário compilado** contra vaults temporários, com `$EDITOR` fake para fluxos de editor headless; testes unitários black-box das APIs exportadas da lógica pura; alvo único que roda unit + e2e + coverage + mutação (gremlins); gate de cobertura ≥90% nas packages de lógica pura (sem gate global de %).

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] `mt --help` imprime ajuda e sai com exit 0
- [ ] Cenário Gherkin de smoke roda verde via binário compilado (não via função chamada)
- [ ] Helper de vault temporário e `$EDITOR` fake disponíveis para todos os cenários
- [ ] Um alvo de automação roda unit + e2e + coverage + mutação
- [ ] Gate de cobertura ≥90% configurado sobre as packages de lógica pura
- [ ] Convenção de exit codes (0 sucesso / 1 erro de usuário / 2 erro de uso) estabelecida e respeitada no smoke
