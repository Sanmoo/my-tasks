Feature: Undefer issues

  mt undefer archives the deferral reminder: without an ID it sweeps the
  vault clearing deferred_until from every expired deferral (printing
  one "Undeferred <id> (was <datetime>)" line per Issue); with an ID it
  clears one Issue even when its deferral is still in the future. Only
  the deferred_until field is touched. These scenarios cover the
  process: the compiled binary against a temporary Vault.

  Background:
    When I run `mt init --prefix pkm <vault>`
    Then the exit code is 0

  Scenario: batch undefer clears every expired deferral and prints what it did
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: expired
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 2
      deferred_until: 2000-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: still future
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
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: expired too
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      deferred_until: 2000-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt undefer --vault <vault>`
    Then the exit code is 0
    And stdout matches "(?s)Undeferred pkm-001.*Undeferred pkm-003"
    And stdout contains "Undeferred pkm-001 (was 2000-01-01T00:00)"
    And stdout contains "Undeferred pkm-003 (was 2000-01-01T00:00)"
    And stdout does not contain "pkm-002"
    And the file "<vault>/issues/pkm-001.md" does not contain "deferred_until:"
    And the file "<vault>/issues/pkm-003.md" does not contain "deferred_until:"
    And the file "<vault>/issues/pkm-002.md" contains "deferred_until: 2999-01-01T00:00"

  Scenario: batch undefer with nothing expired is silent and succeeds
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: no deferral
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt undefer --vault <vault>`
    Then the exit code is 0
    And stdout matches "^$"

  Scenario: undefer with an ID clears a still-future deferral
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: changed my mind
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 2
      deferred_until: 2999-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt undefer --vault <vault> pkm-001`
    Then the exit code is 0
    And stdout contains "Undeferred pkm-001 (was 2999-01-01T00:00)"
    And the file "<vault>/issues/pkm-001.md" does not contain "deferred_until:"

  Scenario: undefer with an ID on an Issue without deferred_until fails
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: no deferral
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt undefer --vault <vault> pkm-001`
    Then the exit code is 1
    And stderr contains "no deferred_until to undefer"
    And the file "<vault>/issues/pkm-001.md" exists

  Scenario: undefer leaves Status and Rank untouched
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: in progress expired
      status: in_progress
      labels: []
      created_at: 2026-01-01T10:00
      rank: 2
      deferred_until: 2000-01-01T00:00
      deadline: 2999-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt undefer --vault <vault> pkm-001`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-001.md" contains "status: in_progress"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-001.md" contains "deadline: 2999-01-01T00:00"
    And the file "<vault>/issues/pkm-001.md" does not contain "deferred_until:"

  Scenario: undefer takes at most one issue ID
    When I run `mt undefer --vault <vault> pkm-001 pkm-002`
    Then the exit code is 2
    And stderr contains "undefer takes at most one issue ID"
