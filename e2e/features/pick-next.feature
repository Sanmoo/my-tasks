Feature: Pick the next Issue

  mt pick-next chooses an available open Issue in Rank order, starts it
  with a started_at timestamp, skips future-deferred Issues, rejects duplicate
  ranks, and allows multiple Issues to remain in_progress.

  Background:
    When I run `mt init --prefix pkm <vault>`
    Then the exit code is 0

  Scenario: pick-next chooses the lowest ranked available open Issue
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: rank two
      status: open
      labels: []
      created_at: 2026-08-15T10:00
      rank: 2
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: rank one
      status: open
      labels: []
      created_at: 2026-08-15T11:00
      rank: 1
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: oldest backlog
      status: open
      labels: []
      created_at: 2020-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt pick-next --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-002 is now in_progress: rank one"
    And the file "<vault>/issues/pkm-002.md" contains "status: in_progress"
    And the file "<vault>/issues/pkm-002.md" matches "started_at: [0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}"
    And the file "<vault>/issues/pkm-001.md" contains "status: open"

  Scenario: pick-next falls back to the oldest Backlog Issue
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: deferred old backlog
      status: open
      labels: []
      created_at: 2020-01-01T10:00
      deferred_until: 2999-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: backlog by id
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-010.md" is written with:
      """
      ---
      title: backlog other id
      status: open
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
      title: newer backlog
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt pick-next --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-002 is now in_progress: backlog by id"
    And the file "<vault>/issues/pkm-002.md" contains "status: in_progress"
    And the file "<vault>/issues/pkm-010.md" contains "status: open"

  Scenario: pick-next skips a future-deferred ranked Issue
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: deferred first rank
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 1
      deferred_until: 2999-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: available second rank
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      rank: 2
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt pick-next --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-002 is now in_progress: available second rank"
    And the file "<vault>/issues/pkm-001.md" contains "status: open"
    And the file "<vault>/issues/pkm-002.md" contains "status: in_progress"

  Scenario: pick-next fails when nothing is available
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: already in progress
      status: in_progress
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
      title: completed
      status: done
      labels: []
      created_at: 2026-01-02T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: deferred
      status: open
      labels: []
      created_at: 2026-01-03T10:00
      deferred_until: 2999-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt pick-next --vault <vault>`
    Then the exit code is 1
    And stderr contains "no available open issues"

  Scenario: pick-next rejects duplicate ranks in the vault
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
      status: in_progress
      labels: []
      created_at: 2026-01-02T10:00
      rank: 1
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt pick-next --vault <vault>`
    Then the exit code is 1
    And stderr contains "duplicate rank: 1"
    And the file "<vault>/issues/pkm-001.md" contains "status: open"

  Scenario: pick-next permits multiple in-progress Issues
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: already in progress
      status: in_progress
      labels: []
      created_at: 2026-01-01T10:00
      started_at: 2026-01-01T11:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: next issue
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      rank: 1
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt pick-next --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-002 is now in_progress: next issue"
    And the file "<vault>/issues/pkm-001.md" contains "status: in_progress"
    And the file "<vault>/issues/pkm-002.md" contains "status: in_progress"
