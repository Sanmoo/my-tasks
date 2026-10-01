# AGENTS.md

Projeto `my-tasks2`: issue tracker pessoal git-friendly (`mt`), uma issue por arquivo Markdown com YAML frontmatter.

## Worktree workflow

Todo trabalho roda num worktree isolado. A `main` só avança por merge fast-forward ou pelo merge de uma PR — nunca por commit direto.

**Julgue antes de começar:** a alteração é simples ou não?

- **Simples** — documentação (inclusive este `AGENTS.md`), ajuste de texto, correção pontual, sem mudança de comportamento. Integre direto na `main` pela sequência **Finish simples**; sem PR.
- **Não simples** — qualquer mudança de comportamento, código, testes ou contrato público, ou que atravesse vários módulos. Abra PR pela sequência **Finish com PR**. A entrega só está concluída depois que o usuário sinalizar que a PR foi mergeada e o worktree e a branch forem removidos.

Na dúvida, trate como não simples.

Start — no checkout principal, com `main` atual:

    git worktree add .worktrees/<slug> -b <slug>

`<slug>` é o nome do arquivo do ticket (ex.: `03-issue-create-show-edit`). Commits e `make check` acontecem dentro de `.worktrees/<slug>`.

### Finish simples — integra direto na `main`

    git rebase main                     # no worktree — resolve conflitos lá
    git merge --ff-only <slug>          # no checkout principal
    git worktree remove .worktrees/<slug>
    git branch -d <slug>

O rebase garante o fast-forward; se o `--ff-only` falhar, volte ao rebase — nunca force.

### Finish com PR — só limpa depois do aval do usuário

    git rebase main                     # no worktree — resolve conflitos lá
    git push -u origin <slug>
    gh pr create --fill --base main

Pare aqui: informe o link da PR e **não** mergeie localmente, **não** remova o worktree nem a branch — a `main` é atualizada pelo merge da PR no GitHub.

Quando o usuário sinalizar que a PR foi mergeada:

    git fetch origin
    git switch main
    git pull --ff-only origin main
    git worktree remove .worktrees/<slug>
    git branch -D <slug>

Se a PR foi mergeada por squash ou rebase, a tip não é ancestral da `main`: confirme a integração por conteúdo (sem diff entre a branch e a `main` atualizada) e só então use `git branch -D`. Com a PR ainda aberta ou o worktree sujo, preserve tudo e pergunte.

Worktrees paralelos convivem: o rebase do finish absorve o que entrou na `main` enquanto isso. Confira o fim com `git worktree list`: o worktree deste ticket deve ter sumido; sem tickets paralelos, resta apenas o checkout principal.

## Agent skills

### Issue tracker

Issues e specs deste repo vivem no GitHub Issues (`Sanmoo/my-tasks`); tickets anteriores à migração ficam em `.scratch/` como histórico. See `docs/agents/issue-tracker.md`.

### Triage labels

Vocabulário padrão de triagem (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Layout single-context: `CONTEXT.md` na raiz + `docs/adr/`. See `docs/agents/domain.md`.
