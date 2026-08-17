Feature: Bare mt lists in_progress

  Invoking mt with no command (after extracting @bookmark) runs the
  resolved vault's in_progress listing — strictly the output of
  `mt list --status in_progress`: the same lines (glyph, ID, title),
  the same [blocked]/[defer ...] suffixes, and the same duplicate-rank
  warning on stderr, computed over the whole vault. With no in_progress
  issues it prints nothing and exits 0; with no resolvable vault it
  fails (exit 1) with the resolution instructions. The list flags do
  not bubble up to the root, and `mt help` / `mt --help` remain the way
  to the root help.

  Scenario: bare mt with a resolved vault shows only in_progress issues
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
    When I run `mt --vault <vault>`
    Then the exit code is 0
    And stdout contains "◐ pkm-002"
    And stdout does not contain "pkm-001"
    And stdout does not contain "pkm-003"

  Scenario: bare mt output matches mt list --status in_progress, suffixes and stderr warning included
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: deferred progress
      status: in_progress
      labels: []
      created_at: 2026-01-01T10:00
      deferred_until: 2999-06-15T14:30
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: blocked progress
      status: in_progress
      labels: []
      created_at: 2026-01-01T10:00
      blocked_by: [pkm-003]
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: open blocker
      status: open
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
      title: open dup rank a
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 1
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-005.md" is written with:
      """
      ---
      title: open dup rank b
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 1
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-001  deferred progress [defer 06-15 14:30]"
    And stdout contains "pkm-002  blocked progress [blocked]"
    And stdout does not contain "pkm-003"
    And stderr contains "Warning: duplicate rank: 1"

  Scenario: bare mt with an empty vault prints nothing and exits zero
    When I run `mt --vault <vault>`
    Then the exit code is 0
    And stdout is empty

  Scenario: bare mt without a resolvable vault fails with instructions
    When I run `mt`
    Then the exit code is 1
    And stderr contains "@bookmark"
    And stderr contains "--vault"
    And stderr contains "default"

  Scenario: bare mt with a @bookmark resolves through it
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: progress via bookmark
      status: in_progress
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<base>/config/mt/config.yaml" is written with:
      """
      default:
      bookmarks:
        bjd: <vault>
      """
    When I run `mt @bjd`
    Then the exit code is 0
    And stdout contains "◐ pkm-001"

  Scenario: bare mt uses the default bookmark when none is given
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: progress via default
      status: in_progress
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<base>/config/mt/config.yaml" is written with:
      """
      default: bjd
      bookmarks:
        bjd: <vault>
      """
    When I run `mt`
    Then the exit code is 0
    And stdout contains "◐ pkm-001"

  Scenario: list flags do not bubble up to bare mt
    When I run `mt --status open`
    Then the exit code is 2
    And stderr contains "unknown flag"
    And stdout does not contain "Usage:"

  Scenario: mt help still shows the root help
    When I run `mt help`
    Then the exit code is 0
    And stdout contains "Usage:"
    And stdout contains "Exit codes"
