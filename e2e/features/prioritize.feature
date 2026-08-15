Feature: Prioritize via $EDITOR

  mt prioritize opens $EDITOR on a [P]/[ ] buffer of the vault's open and
  in_progress issues, then applies the saved order: reordered [P] lines
  change the ranks, [ ]↔[P] toggles move an issue between the queue and
  the Backlog, and the ranks are renormalized 1..N with only the changed
  files rewritten. An invalid buffer is rejected and nothing is applied.
  The buffer/parse/plan decisions are pure logic covered at Seam 2
  (internal/priority); these scenarios cover the process: the compiled
  binary against a temporary Vault with a fake $EDITOR.

  Scenario: prioritize without a vault fails with instructions
    When I run `mt prioritize`
    Then the exit code is 1
    And stderr contains "@bookmark"
    And stderr contains "--vault"
    And stderr contains "default"

  Scenario: reordering [P] lines changes the order and renormalizes 1..N
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
      title: rank one
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      rank: 1
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: rank nine
      status: in_progress
      labels: []
      created_at: 2026-01-03T10:00
      rank: 9
      ---

      ## Description
      ## Notes
      ## Comments
      """
    Given the fake editor writes
      """
      [P] pkm-003  rank nine
      [P] pkm-001  rank five
      [P] pkm-002  rank one
      """
    When I run `mt prioritize --vault <vault>`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-003.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 3"
    And the file "<vault>/issues/pkm-001.md" does not contain "rank: 5"
    And the file "<vault>/issues/pkm-003.md" does not contain "rank: 9"

  Scenario: toggling [ ]↔[P] moves issues between the queue and the Backlog
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
      rank: 2
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
    Given the fake editor writes
      """
      [P] pkm-003  backlog
      [ ] pkm-002  second
      [P] pkm-001  first
      """
    When I run `mt prioritize --vault <vault>`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-003.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-002.md" does not contain "rank:"

  Scenario: open and in_progress are included, done is left untouched
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: open one
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
      title: progress two
      status: in_progress
      labels: []
      created_at: 2026-01-02T10:00
      rank: 2
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
      created_at: 2026-01-03T10:00
      rank: 3
      ---

      ## Description
      ## Notes
      ## Comments
      """
    Given the fake editor writes
      """
      [P] pkm-002  progress two
      [P] pkm-001  open one
      """
    When I run `mt prioritize --vault <vault>`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-002.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-003.md" contains "status: done"
    And the file "<vault>/issues/pkm-003.md" contains "rank: 3"

  Scenario: only the files whose rank changed are rewritten
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      # keep me — a rewrite would drop this comment
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
      rank: 2
      ---

      ## Description
      ## Notes
      ## Comments
      """
    Given the fake editor writes
      """
      [P] pkm-001  first
      [P] pkm-002  second
      """
    When I run `mt prioritize --vault <vault>`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-001.md" contains "# keep me"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 2"

  Scenario: an unknown issue ID is rejected and nothing is applied
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
    Given the fake editor writes
      """
      [P] pkm-001  first
      [P] pkm-999  ghost
      """
    When I run `mt prioritize --vault <vault>`
    Then the exit code is 1
    And stderr contains "unknown issue ID"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"

  Scenario: a duplicated issue ID is rejected and nothing is applied
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
    Given the fake editor writes
      """
      [P] pkm-001  first
      [P] pkm-001  again
      """
    When I run `mt prioritize --vault <vault>`
    Then the exit code is 1
    And stderr contains "duplicate"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"

  Scenario: an issue missing from the buffer is rejected and nothing is applied
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
      rank: 2
      ---

      ## Description
      ## Notes
      ## Comments
      """
    Given the fake editor writes
      """
      [P] pkm-001  first
      """
    When I run `mt prioritize --vault <vault>`
    Then the exit code is 1
    And stderr contains "missing"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 2"
