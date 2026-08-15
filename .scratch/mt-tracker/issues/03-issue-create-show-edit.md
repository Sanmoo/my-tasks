# 03 — Criar, ver e editar Issues (schema completo)

**What to build:** `mt create <título>` e `mt q <título>` (imprime só o ID) gravam o arquivo da Issue com o frontmatter exato da spec: campos sempre presentes (`title`, `status`, `labels`, `created_at`), demais campos presentes só quando têm valor, datetime `YYYY-MM-DDTHH:MM` naive, corpo com seções `Description`/`Notes`/`Comments`, sem campo `id` (o nome do arquivo é a autoridade) e sem `updated_at`. `created_at` é carimbado automaticamente; IDs usam o prefixo do vault + sufixo curto aleatório sem colisão; labels livres na criação. `mt show` lê a Issue de volta; `mt edit` abre o `$EDITOR` preservando o que não foi editado.

**Blocked by:** 02 — Vault init, config e resolução de vault

**Status:** ready-for-agent

- [ ] `create` grava arquivo com o schema exato da spec
- [ ] `q` imprime apenas o ID
- [ ] `created_at` automático no formato `YYYY-MM-DDTHH:MM` naive
- [ ] IDs únicos com o prefixo do vault
- [ ] Labels livres aceitas na criação
- [ ] `show` exibe frontmatter + corpo
- [ ] `edit` abre o $EDITOR e o round-trip preserva o conteúdo não editado
- [ ] Unit (Seam 2): round-trip do frontmatter (ordem estável, campos opcionais somente-quando-setados) com alta cobertura
