Feature: Check vault

  mt check audits Issue frontmatter, status values and Rank integrity;
  --fix repairs the ranked queue by renormalizing Ranks to 1..N.

  Background:
    When I run `mt init --prefix pkm <vault>`
    Then the exit code is 0

  Scenario: check without a Vault fails with instructions
    When I run `mt check`
    Then the exit code is 1
    And stderr contains "@bookmark"
    And stderr contains "--vault"
    And stderr contains "default"

  Scenario: duplicate Ranks are an error
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
      created_at: 2026-01-02T10:00
      rank: 1
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "duplicate rank"
    And stderr contains "1"

  Scenario: a Rank gap is a soft warning
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
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: third
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      rank: 3
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 0
    And stderr contains "rank gap"
    And stderr contains "2"

  Scenario: malformed YAML frontmatter is an error
    Given the file "<vault>/issues/pkm-bad.md" is written with:
      """
      ---
      title: broken
      status: open
      labels: [unclosed
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "frontmatter"
    And stderr contains "pkm-bad"

  Scenario: incomplete frontmatter is an error
    Given the file "<vault>/issues/pkm-incomplete.md" is written with:
      """
      ---
      title: incomplete
      status: open
      labels: []
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "malformed frontmatter"
    And stderr contains "created_at"

  Scenario: forbidden frontmatter fields are an error
    Given the file "<vault>/issues/pkm-forbidden.md" is written with:
      """
      ---
      title: forbidden fields
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      id: pkm-forbidden
      updated_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "unknown field"
    And stderr contains "id"

  Scenario: a fractional Rank is malformed
    Given the file "<vault>/issues/pkm-fraction.md" is written with:
      """
      ---
      title: fractional rank
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 1.2
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "rank must be an integer"
    And stderr contains "pkm-fraction"

  Scenario: a status outside the Vault configuration is an error
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: blocked issue
      status: blocked
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "status"
    And stderr contains "blocked"
    And stderr contains "open, in_progress, done"

  Scenario: an invalid datetime format is an error
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: bad date
      status: open
      labels: []
      created_at: 2026-01-01
      deadline: not-a-datetime
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "datetime"
    And stderr contains "created_at"
    And stderr contains "pkm-001"

  Scenario: an optional datetime field is also validated
    Given the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: bad deferred date
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      deferred_until: 2026-01-01
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "deferred_until"
    And stderr contains "pkm-002"

  Scenario: a non-positive Rank is an error
    Given the file "<vault>/issues/pkm-zero.md" is written with:
      """
      ---
      title: zero rank
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 0
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "invalid rank"
    And stderr contains "0"

  Scenario: --fix renormalizes ranked Issues to 1..N
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: rank five
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 5
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: rank nine
      status: in_progress
      labels: []
      created_at: 2026-01-02T10:00
      rank: 9
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: backlog
      status: open
      labels: []
      created_at: 2026-01-03T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --fix --vault <vault>`
    Then the exit code is 0
    And stdout contains "Fixed"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-001.md" does not contain "rank: 5"
    And the file "<vault>/issues/pkm-002.md" does not contain "rank: 9"
    And the file "<vault>/issues/pkm-003.md" does not contain "rank:"
