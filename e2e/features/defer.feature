Feature: Defer issues

  mt defer <id> <when> sets deferred_until keeping the Issue open: it
  accepts an absolute YY-MM-DD HH:MM (the hour is kept, not truncated)
  and relative durations (+2d, +1w, +3h) computed from now. A deferred
  Issue disappears from list until now >= deferred_until, then reappears
  on its own; when the deferral is expired, mt undefer archives the
  reminder. The datetime parsing is
  decision-dense pure logic covered at Seam 2 (internal/deferral); these
  scenarios cover the process: the compiled binary against a temporary
  Vault.

  Background:
    When I run `mt init --prefix pkm <vault>`
    Then the exit code is 0

  Scenario: defer with an absolute datetime keeps the hour
    When I run `mt create --vault <vault> "adiar compra"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt defer --vault <vault> <id> 26-08-20 08:00`
    Then the exit code is 0
    And stdout contains "<id> deferred until 2026-08-20T08:00"
    And the file "<vault>/issues/<id>.md" contains "deferred_until: 2026-08-20T08:00"
    And the file "<vault>/issues/<id>.md" contains "status: open"

  Scenario: defer with a relative duration computes from now
    When I run `mt create --vault <vault> "adiar compra"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt defer --vault <vault> <id> +2d`
    Then the exit code is 0
    And stdout contains "<id> deferred until "
    And the file "<vault>/issues/<id>.md" matches "deferred_until: [0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}"
    And the file "<vault>/issues/<id>.md" contains "status: open"

  Scenario: deferring an in-progress Issue returns it to open
    When I run `mt create --vault <vault> "adiar trabalho em andamento"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt status --vault <vault> <id> in_progress`
    Then the exit code is 0
    When I run `mt defer --vault <vault> <id> 99-01-01 00:00`
    Then the exit code is 0
    And the file "<vault>/issues/<id>.md" contains "status: open"
    And the file "<vault>/issues/<id>.md" contains "deferred_until: 2099-01-01T00:00"

  Scenario: a deferred issue hides from list until its time arrives
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: soon
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
      title: later
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt defer --vault <vault> pkm-001 20-01-01 00:00`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-001.md" contains "deferred_until: 2020-01-01T00:00"
    And the file "<vault>/issues/pkm-001.md" contains "status: open"
    When I run `mt defer --vault <vault> pkm-002 99-01-01 00:00`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-002.md" contains "deferred_until: 2099-01-01T00:00"
    And the file "<vault>/issues/pkm-002.md" contains "status: open"
    When I run `mt list --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-001"
    And stdout does not contain "pkm-002"
    When I run `mt list --vault <vault> --all`
    Then the exit code is 0
    And stdout contains "pkm-001"
    And stdout contains "pkm-002  later [defer 01-01 00:00]"

  Scenario: defer rejects a malformed time
    When I run `mt create --vault <vault> t`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt defer --vault <vault> <id> not-a-time`
    Then the exit code is 2
    And stderr contains "YY-MM-DD HH:MM"
    And the file "<vault>/issues/<id>.md" does not contain "deferred_until:"

  Scenario: defer without a time is a usage error
    When I run `mt create --vault <vault> t`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt defer --vault <vault> <id>`
    Then the exit code is 2
    And stderr contains "defer needs an issue ID and a time"
