Feature: Status transitions

  done (alias close) closes an Issue stamping completed_at; reopen
  returns it to open clearing completed_at and started_at; status
  transitions freely between statuses, validating against the vault's
  configured list. The transition rules themselves are decision-dense
  pure logic covered at Seam 2 (internal/issue, internal/vault); these
  scenarios cover the process: the compiled binary against a temporary
  Vault.

  Background:
    When I run `mt init --prefix pkm <vault>`
    Then the exit code is 0

  Scenario: done closes the issue and stamps completed_at
    When I run `mt create --vault <vault> "comprar material"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt done --vault <vault> <id>`
    Then the exit code is 0
    And stdout contains "<id> is now done: comprar material"
    And the file "<vault>/issues/<id>.md" contains "status: done"
    And the file "<vault>/issues/<id>.md" matches "completed_at: [0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}"

  Scenario: close is an alias of done
    When I run `mt create --vault <vault> "comprar material"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt close --vault <vault> <id>`
    Then the exit code is 0
    And stdout contains "<id> is now done: comprar material"
    And the file "<vault>/issues/<id>.md" contains "status: done"
    And the file "<vault>/issues/<id>.md" contains "completed_at:"

  Scenario: reopen clears completed_at and started_at and returns to open
    Given the file "<vault>/issues/pkm-0001.md" is written with:
      """
      ---
      title: t
      status: done
      labels: []
      created_at: 2026-08-15T09:30
      started_at: 2026-08-20T09:00
      completed_at: 2026-08-21T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt reopen --vault <vault> pkm-0001`
    Then the exit code is 0
    And stdout contains "pkm-0001 is now open: t"
    And the file "<vault>/issues/pkm-0001.md" contains "status: open"
    And the file "<vault>/issues/pkm-0001.md" does not contain "completed_at:"
    And the file "<vault>/issues/pkm-0001.md" does not contain "started_at:"

  Scenario: status accepts a configured status
    When I run `mt create --vault <vault> t`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt status --vault <vault> <id> in_progress`
    Then the exit code is 0
    And stdout contains "<id> is now in_progress: t"
    And the file "<vault>/issues/<id>.md" contains "status: in_progress"

  Scenario: status with an empty or blank title omits the title suffix
    Given the file "<vault>/issues/pkm-0001.md" is written with:
      """
      ---
      title: "   "
      status: open
      labels: []
      created_at: 2026-08-15T09:30
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt status --vault <vault> pkm-0001 in_progress`
    Then the exit code is 0
    And stdout contains "pkm-0001 is now in_progress"
    And stdout does not contain "is now in_progress: "
    And the file "<vault>/issues/pkm-0001.md" contains "status: in_progress"

  Scenario: status flattens newlines in the title
    Given the file "<vault>/issues/pkm-0001.md" is written with:
      """
      ---
      title: "linha um\nlinha dois"
      status: open
      labels: []
      created_at: 2026-08-15T09:30
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt status --vault <vault> pkm-0001 in_progress`
    Then the exit code is 0
    And stdout contains "pkm-0001 is now in_progress: linha um linha dois"
    And the file "<vault>/issues/pkm-0001.md" contains "status: in_progress"

  Scenario: status accepts a custom status from the vault config
    Given the file "<vault>/mt.yaml" is written with:
      """
      prefix: pkm
      status: [open, in_progress, done, blocked]
      """
    When I run `mt create --vault <vault> t`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt status --vault <vault> <id> blocked`
    Then the exit code is 0
    And the file "<vault>/issues/<id>.md" contains "status: blocked"

  Scenario: status rejects a value outside the config with a clear error
    When I run `mt create --vault <vault> t`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt status --vault <vault> <id> bogus`
    Then the exit code is 1
    And stderr contains "bogus"
    And stderr contains "open, in_progress, done"
    And the file "<vault>/issues/<id>.md" contains "status: open"

  Scenario: status transitions freely with no state machine
    When I run `mt create --vault <vault> t`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt status --vault <vault> <id> done`
    Then the exit code is 0
    When I run `mt status --vault <vault> <id> in_progress`
    Then the exit code is 0
    And the file "<vault>/issues/<id>.md" contains "status: in_progress"
    When I run `mt status --vault <vault> <id> done`
    Then the exit code is 0
    And the file "<vault>/issues/<id>.md" contains "status: done"
    And the file "<vault>/issues/<id>.md" does not contain "completed_at:"
