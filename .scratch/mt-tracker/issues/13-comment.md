# 13 — Comentários

**What to build:** `mt comment <id> <texto>` anexa um comentário à seção `Comments` da Issue com heading de timestamp e âncora estável por comentário. Append-only: o corpo existente é preservado intacto.

**Blocked by:** 03 — Criar, ver e editar Issues (schema completo)

**Status:** ready-for-agent

- [ ] Comentário anexado com heading de timestamp
- [ ] Cada comentário recebe âncora estável
- [ ] Corpo existente preservado intacto
- [ ] Sem argumento de texto → erro de uso (exit 2)
