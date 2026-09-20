# 01 — undefer line shows the Issue title

**What to build:** `mt undefer`, in both modes (batch without arguments and per-id), ends each `Undeferred` confirmation line with the Issue's title in the house style: `Undeferred <id> (was <datetime>): <título>`. Title whitespace collapses to single spaces so the line never wraps; a blank/whitespace-only title keeps the line exactly as today, without a dangling `: `. The `(was <datetime>)` block keeps its position and raw value. Nothing else changes — exit codes, silent batch, Status/Rank untouched, and `mt overdue`/`ready`/`pick-next` stay as they are. The title-suffix collapse logic is shared with the transition confirmation line rather than copied. The README's `mt undefer` section and output example show the new shape.

**Blocked by:** None — can start immediately.

**Status:** claimed

- [ ] Batch undefer prints `Undeferred <id> (was <datetime>): <título>` per expired deferral
- [ ] Per-id undefer prints the same line shape
- [ ] Title whitespace collapses to single spaces (line stays on one line)
- [ ] Blank/whitespace-only title prints the line without the `: ` suffix
- [ ] `(was <datetime>)` keeps its position and raw stored value
- [ ] E2E scenarios updated, plus new ones for whitespace-collapse and blank title — full `make check` green
- [ ] README `mt undefer` section and example show the title
- [ ] Exit codes, silent-empty-batch behavior, and the other commands' output are unchanged
