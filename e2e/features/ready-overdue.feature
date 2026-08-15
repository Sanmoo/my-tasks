Feature: Ready and overdue issues

  mt ready lists open Issues that are available now, in priority order.
  mt overdue lists non-done Issues whose Deadline has passed. Both commands
  use the CLI-process seam against a temporary Vault and succeed silently
  when their result is empty.

  Scenario: ready lists only available open issues in priority order
    Given the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: second ready
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      rank: 2
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: first ready
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      rank: 1
      deadline: 2000-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: backlog ready
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
      title: future deferred
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      deferred_until: 2999-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-005.md" is written with:
      """
      ---
      title: in progress
      status: in_progress
      labels: []
      created_at: 2026-01-01T10:00
      rank: 3
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt ready --vault <vault>`
    Then the exit code is 0
    And stdout matches "(?s)pkm-001.*pkm-002.*pkm-003"
    And stdout does not contain "pkm-004"
    And stdout does not contain "pkm-005"

  Scenario: overdue lists every non-done Issue whose Deadline has passed
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: open overdue
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      deadline: 2000-01-01T00:00
      deferred_until: 2999-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: progress overdue
      status: in_progress
      labels: []
      created_at: 2026-01-01T10:00
      deadline: 2000-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-003.md" is written with:
      """
      ---
      title: done overdue
      status: done
      labels: []
      created_at: 2026-01-01T10:00
      deadline: 2000-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-004.md" is written with:
      """
      ---
      title: future deadline
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      deadline: 2999-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt overdue --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-001"
    And stdout contains "pkm-002"
    And stdout does not contain "pkm-003"
    And stdout does not contain "pkm-004"

  Scenario: ready and overdue succeed silently when nothing applies
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: done
      status: done
      labels: []
      created_at: 2026-01-01T10:00
      deadline: 2000-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt ready --vault <vault>`
    Then the exit code is 0
    And stdout matches "^$"
    When I run `mt overdue --vault <vault>`
    Then the exit code is 0
    And stdout matches "^$"
