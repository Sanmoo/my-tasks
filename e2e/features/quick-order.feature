Feature: Quick order commands

  mt top, bottom, rank and unrank adjust the queue without opening $EDITOR.
  They reuse the priority renormalization and rewrite only changed Issues.

  Scenario: top promotes an Issue and renormalizes the queue
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
      title: third
      status: open
      labels: []
      created_at: 2026-01-03T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt top --vault <vault> pkm-003`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-003.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 3"

  Scenario: bottom moves a queued Issue to the end
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
    When I run `mt bottom --vault <vault> pkm-001`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-002.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-003.md" does not contain "rank:"

  Scenario: rank inserts a Backlog Issue and shifts later ranks
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
      title: third
      status: in_progress
      labels: []
      created_at: 2026-01-03T10:00
      rank: 3
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-004.md" is written with:
      """
      ---
      title: backlog
      status: open
      labels: []
      created_at: 2026-01-04T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt rank --vault <vault> pkm-004 2`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-004.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 3"
    And the file "<vault>/issues/pkm-003.md" contains "rank: 4"

  Scenario: unrank returns an Issue to the Backlog
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
      rank: 2
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
      created_at: 2026-01-03T10:00
      rank: 3
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt unrank --vault <vault> pkm-002`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-003.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-002.md" does not contain "rank:"

  Scenario: quick ordering leaves unchanged files untouched
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
      title: third
      status: open
      labels: []
      created_at: 2026-01-03T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-004.md" is written with:
      """
      ---
      # keep me — a rewrite would remove this comment
      title: untouched
      status: open
      labels: []
      created_at: 2026-01-04T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt top --vault <vault> pkm-003`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-004.md" contains "# keep me"
    And the file "<vault>/issues/pkm-003.md" contains "rank: 1"
