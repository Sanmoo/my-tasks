# 02 — Vault init, config e resolução de vault

**What to build:** `mt init [dir]` cria um Vault utilizável (diretório de issues + config do vault com prefixo de ID e lista de Status); a config global é detectada automaticamente (XDG) com bookmarks e favorito principal; resolução de vault na precedência `@bookmark` > `--vault <path>` > favorito principal; sem nenhum deles, o comando falha em stderr com instruções (exit 1). Status default (`open`, `in_progress`, `done`) valem quando a config não os define.

**Blocked by:** 01 — Esqueleto do projeto e seams de teste

**Status:** ready-for-agent

- [ ] `mt init` cria vault utilizável com config default
- [ ] init aceita prefixo de ID e lista de status customizada
- [ ] Config global com bookmarks e favorito principal é lida automaticamente
- [ ] Resolução respeita a precedência `@` > `--vault` > `default`
- [ ] Sem nenhum definido, qualquer comando falha com mensagem de instrução e exit 1
- [ ] Unit (Seam 2): precedência de resolução coberta com alta cobertura
