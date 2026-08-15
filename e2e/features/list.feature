Feature: List issues

  mt list shows the vault's issues in priority order — ranked first
  (lowest rank first), then the Backlog by created_at, then ID — with a
  status glyph per line. done and future-deferred issues are hidden by
  default (--all reveals them, marking future-deferred with a
  [defer MM-DD HH:MM] suffix); --status and --label narrow the view; a
  duplicate rank prints a warning to stderr. The ordering comparator is
  decision-dense pure logic covered at Seam 2 (internal/list); these
  scenarios cover the process: the compiled binary against a temporary
  Vault.

  Scenario: list orders by rank then backlog created_at then id
    Given the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: rank two
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 2
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-004.md" is written with:
      """
      ---
      title: backlog newer
      status: open
      labels: []
      created_at: 2026-01-03T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: rank one
      status: open
      labels: []
      created_at: 2026-01-05T10:00
      rank: 1
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: backlog older
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt list --vault <vault>`
    Then the exit code is 0
    And stdout matches "(?s)○ pkm-001\s+rank one.*○ pkm-002\s+rank two.*○ pkm-003\s+backlog older.*○ pkm-004\s+backlog newer"

  Scenario: list shows the status glyph and a fallback for custom statuses
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: open one
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: progress two
      status: in_progress
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: done three
      status: done
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-004.md" is written with:
      """
      ---
      title: blocked four
      status: blocked
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt list --vault <vault> --all`
    Then the exit code is 0
    And stdout contains "○ pkm-001"
    And stdout contains "◐ pkm-002"
    And stdout contains "● pkm-003"
    And stdout contains "? pkm-004"

  Scenario: done issues are hidden by default and shown with --all
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: alive
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: closed
      status: done
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt list --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-001"
    And stdout does not contain "pkm-002"
    When I run `mt list --vault <vault> --all`
    Then the exit code is 0
    And stdout contains "pkm-002"

  Scenario: future-deferred issues are hidden by default and marked via --all
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: available
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: deferred future
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      deferred_until: 2999-06-15T14:30
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: deferred past
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      deferred_until: 2000-03-10T09:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt list --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-001"
    And stdout does not contain "pkm-002"
    And stdout contains "pkm-003"
    When I run `mt list --vault <vault> --all`
    Then the exit code is 0
    And stdout contains "pkm-002  deferred future [defer 06-15 14:30]"
    And stdout does not contain "pkm-003  deferred past [defer"

  Scenario: list filters by status
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: open one
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: progress two
      status: in_progress
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: done three
      status: done
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt list --vault <vault> --status in_progress`
    Then the exit code is 0
    And stdout contains "pkm-002"
    And stdout does not contain "pkm-001"
    And stdout does not contain "pkm-003"
    When I run `mt list --vault <vault> --status done`
    Then the exit code is 0
    And stdout contains "pkm-003"
    And stdout does not contain "pkm-001"

  Scenario: list filters by label
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: compras issue
      status: open
      labels: [compras, familia]
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: saude issue
      status: open
      labels: [saude]
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt list --vault <vault> --label compras`
    Then the exit code is 0
    And stdout contains "pkm-001"
    And stdout does not contain "pkm-002"
    When I run `mt list --vault <vault> --label saude`
    Then the exit code is 0
    And stdout contains "pkm-002"
    And stdout does not contain "pkm-001"
    When I run `mt list --vault <vault> --label compras --label saude`
    Then the exit code is 0
    And stdout contains "pkm-001"
    And stdout contains "pkm-002"

  Scenario: list warns about a duplicate rank
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: first
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 1
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: second
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 1
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt list --vault <vault>`
    Then the exit code is 0
    And stderr contains "Warning: duplicate rank: 1"
    And stdout contains "pkm-001"
    And stdout contains "pkm-002"

  Scenario: list without a vault fails with instructions
    When I run `mt list`
    Then the exit code is 1
    And stderr contains "@bookmark"
    And stderr contains "--vault"
    And stderr contains "default"
