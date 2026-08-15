# AGENTS.md

Projeto `my-tasks2`: issue tracker pessoal git-friendly (`mt`), uma issue por arquivo Markdown com YAML frontmatter.

## Worktree workflow

Todo trabalho de implementação roda num worktree isolado; `main` só avança por merges fast-forward.

Start — no checkout principal, com `main` atual:

    git worktree add .worktrees/<slug> -b <slug>

`<slug>` é o nome do arquivo do ticket (ex.: `03-issue-create-show-edit`). Commits e `make check` acontecem dentro de `.worktrees/<slug>`.

Finish:

    git rebase main                     # no worktree — resolve conflitos lá
    git merge --ff-only <slug>          # no checkout principal
    git worktree remove .worktrees/<slug>
    git branch -d <slug>

O rebase garante o fast-forward; se o `--ff-only` falhar, volte ao rebase — nunca force. Worktrees paralelos convivem: o rebase do finish absorve o que entrou na `main` enquanto isso. Confira o fim com `git worktree list` (apenas o checkout principal).

## Agent skills

### Issue tracker

Issues e specs deste repo vivem como arquivos Markdown em `.scratch/`. See `docs/agents/issue-tracker.md`.

### Triage labels

Vocabulário padrão de triagem (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Layout single-context: `CONTEXT.md` na raiz + `docs/adr/`. See `docs/agents/domain.md`.
