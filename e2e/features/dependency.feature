Feature: Issue dependencies (blocked_by)

  mt dep add <id> <blocker> records that <blocker> blocks <id> in the
  Issue's blocked_by field; mt dep rm removes the reference. An Issue is
  blocked — computed state, not a status — while any of its blockers is
  not done: list marks it [blocked], ready and pick-next skip it, and
  closing the blocker unblocks it on its own. mt check validates the
  references (existence, no self-block, no cycles).

  Background:
    When I run `mt init --prefix pkm <vault>`
    Then the exit code is 0

  Scenario: dep add blocks an Issue until the blocker is done
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: blocker
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
      title: dependent
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      rank: 2
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt dep add --vault <vault> pkm-002 pkm-001`
    Then the exit code is 0
    And stdout contains "pkm-002 is now blocked by pkm-001"
    And the file "<vault>/issues/pkm-002.md" contains "blocked_by: [pkm-001]"
    When I run `mt list --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-002  dependent [blocked]"
    And stdout does not contain "pkm-001  blocker [blocked]"
    When I run `mt ready --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-001"
    And stdout does not contain "pkm-002"
    When I run `mt pick-next --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-001 is now in_progress: blocker"
    And the file "<vault>/issues/pkm-002.md" contains "status: open"
    When I run `mt done --vault <vault> pkm-001`
    Then the exit code is 0
    And stdout contains "pkm-001 is now done: blocker"
    When I run `mt list --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-002  dependent"
    And stdout does not contain "[blocked]"
    When I run `mt pick-next --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-002 is now in_progress: dependent"

  Scenario: a blocker being done unblocks without an explicit operation
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: blocker
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
      title: dependent
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      blocked_by: [pkm-001]
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt ready --vault <vault>`
    Then the exit code is 0
    And stdout does not contain "pkm-002"
    When I run `mt done --vault <vault> pkm-001`
    Then the exit code is 0
    When I run `mt ready --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-002"

  Scenario: pick-next fails when every open Issue is blocked
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: blocked issue
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 1
      blocked_by: [pkm-002]
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: in progress blocker
      status: in_progress
      labels: []
      created_at: 2026-01-02T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt pick-next --vault <vault>`
    Then the exit code is 1
    And stderr contains "no available open issues"
    And the file "<vault>/issues/pkm-001.md" contains "status: open"

  Scenario: dep add rejects an unknown blocker
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: subject
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt dep add --vault <vault> pkm-001 pkm-999`
    Then the exit code is 1
    And stderr contains "issue pkm-999 not found"
    And the file "<vault>/issues/pkm-001.md" does not contain "blocked_by"

  Scenario: dep add rejects a self-block
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: subject
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt dep add --vault <vault> pkm-001 pkm-001`
    Then the exit code is 1
    And stderr contains "cannot block itself"
    And the file "<vault>/issues/pkm-001.md" does not contain "blocked_by"

  Scenario: dep add and dep rm with malformed arguments are usage errors
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: subject
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt dep add --vault <vault> pkm-001`
    Then the exit code is 2
    And stderr contains "dep add needs an issue ID and a blocker ID"
    When I run `mt dep rm --vault <vault> pkm-001 pkm-002 extra`
    Then the exit code is 2
    And stderr contains "dep rm needs an issue ID and a blocker ID"

  Scenario: dep rm removes the blocker and the [blocked] marker
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: blocker
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
      title: dependent
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      blocked_by: [pkm-001]
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt dep rm --vault <vault> pkm-002 pkm-001`
    Then the exit code is 0
    And stdout contains "pkm-002 is no longer blocked by pkm-001"
    And the file "<vault>/issues/pkm-002.md" does not contain "blocked_by"
    When I run `mt list --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-002  dependent"
    And stdout does not contain "[blocked]"

  Scenario: dep rm is idempotent and clears a stale reference
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: subject
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      blocked_by: [pkm-999]
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt dep rm --vault <vault> pkm-001 pkm-999`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-001.md" does not contain "blocked_by"
    When I run `mt dep rm --vault <vault> pkm-001 pkm-999`
    Then the exit code is 0
    And stdout contains "pkm-001 is no longer blocked by pkm-999"

  Scenario: check accepts valid blocked_by references
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: blocker
      status: done
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
      title: dependent
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      blocked_by: [pkm-001]
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 0
    And stdout contains "OK"

  Scenario: check reports a blocked_by reference to an unknown Issue
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: subject
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      blocked_by: [pkm-999]
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "blocked_by of issue pkm-001 references unknown issue pkm-999"

  Scenario: check reports a self-block
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: subject
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      blocked_by: [pkm-001]
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "lists itself in blocked_by"

  Scenario: check reports a blocked_by cycle
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: first
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      blocked_by: [pkm-002]
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
      created_at: 2026-01-02T10:00
      blocked_by: [pkm-001]
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "blocked_by cycle"
    And stderr contains "pkm-001 -> pkm-002 -> pkm-001"

  Scenario: a blocked and future-deferred Issue shows both suffixes in list, with and without --all
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: blocker
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
      title: deferred and blocked
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      deferred_until: 2999-01-01T00:00
      blocked_by: [pkm-001]
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt list --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-002  deferred and blocked [defer 01-01 00:00] [blocked]"
    When I run `mt list --vault <vault> --all`
    Then the exit code is 0
    And stdout contains "pkm-002  deferred and blocked [defer 01-01 00:00] [blocked]"
