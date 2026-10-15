Feature: Claim issues

  mt claim <id> <who> records the claimed_by field: a session assumed
  the Issue before any work started. Claim is a field, not a status —
  the Issue keeps its status — and a set claim removes availability, so
  ready and pick-next skip it until mt unclaim clears the field, reopen
  wipes it, or done archives it as history. The claimer identity is the
  value's only job: concurrency-safe possession in a personal tracker.

  Background:
    When I run `mt init --prefix pkm <vault>`
    Then the exit code is 0

  Scenario: claim records the claimer in the file and on stdout
    When I run `mt create --vault <vault> "reservar antes de começar"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt claim --vault <vault> <id> pi-sessao-7`
    Then the exit code is 0
    And stdout contains "<id> claimed by pi-sessao-7"
    And the file "<vault>/issues/<id>.md" contains "claimed_by: pi-sessao-7"
    And the file "<vault>/issues/<id>.md" contains "status: open"

  Scenario: a claimed Issue is skipped by ready and pick-next
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: claimed first
      status: open
      labels: []
      created_at: 2026-08-15T09:00
      rank: 1
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-002.md" is written with:
      """
      ---
      title: free second
      status: open
      labels: []
      created_at: 2026-08-15T10:00
      rank: 2
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt claim --vault <vault> pkm-001 pi-sessao-7`
    Then the exit code is 0
    When I run `mt ready --vault <vault>`
    Then the exit code is 0
    And stdout does not contain "claimed first"
    And stdout contains "free second"
    When I run `mt pick-next --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-002 is now in_progress: free second"
    And the file "<vault>/issues/pkm-001.md" contains "status: open"

  Scenario: a claimed Issue stays visible in list with a [claimed] suffix
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: held by a session
      status: open
      labels: []
      created_at: 2026-08-15T09:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt claim --vault <vault> pkm-001 pi-sessao-7`
    Then the exit code is 0
    When I run `mt list --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-001  held by a session [claimed]"

  Scenario: unclaim clears the field and returns the Issue to availability
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: released again
      status: open
      labels: []
      created_at: 2026-08-15T09:00
      claimed_by: pi-sessao-7
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt ready --vault <vault>`
    Then the exit code is 0
    And stdout does not contain "released again"
    When I run `mt unclaim --vault <vault> pkm-001`
    Then the exit code is 0
    And stdout contains "pkm-001 unclaimed"
    And the file "<vault>/issues/pkm-001.md" does not contain "claimed_by:"
    When I run `mt ready --vault <vault>`
    Then the exit code is 0
    And stdout contains "released again"

  Scenario: unclaim works on a done Issue (rare manual correction)
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: closed with a stale claim
      status: done
      labels: []
      created_at: 2026-08-15T09:00
      completed_at: 2026-08-16T12:00
      claimed_by: pi-sessao-7
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt unclaim --vault <vault> pkm-001`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-001.md" does not contain "claimed_by:"
    And the file "<vault>/issues/pkm-001.md" contains "status: done"

  Scenario: claim overwrites an existing claim
    When I run `mt create --vault <vault> "transferir posse"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt claim --vault <vault> <id> sessao-a`
    Then the exit code is 0
    When I run `mt claim --vault <vault> <id> sessao-b`
    Then the exit code is 0
    And stdout contains "<id> claimed by sessao-b"
    And the file "<vault>/issues/<id>.md" contains "claimed_by: sessao-b"
    And the file "<vault>/issues/<id>.md" does not contain "claimed_by: sessao-a"

  Scenario: claim is not allowed on a non-open Issue
    When I run `mt create --vault <vault> "já em andamento"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt status --vault <vault> <id> in_progress`
    Then the exit code is 0
    When I run `mt claim --vault <vault> <id> pi-sessao-7`
    Then the exit code is 1
    And stderr contains "open"
    And the file "<vault>/issues/<id>.md" does not contain "claimed_by:"

  Scenario: claim is not allowed on a done Issue
    When I run `mt create --vault <vault> "já fechada"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt done --vault <vault> <id>`
    Then the exit code is 0
    When I run `mt claim --vault <vault> <id> pi-sessao-7`
    Then the exit code is 1
    And the file "<vault>/issues/<id>.md" does not contain "claimed_by:"

  Scenario: claim without a claimer is a usage error
    When I run `mt create --vault <vault> t`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt claim --vault <vault> <id>`
    Then the exit code is 2
    And the file "<vault>/issues/<id>.md" does not contain "claimed_by:"

  Scenario: unclaim on an Issue with no claim is a user error
    When I run `mt create --vault <vault> "nunca claimada"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt unclaim --vault <vault> <id>`
    Then the exit code is 1
    And stderr contains "no claim"

  Scenario: claim on a missing Issue is a user error
    When I run `mt claim --vault <vault> pkm-999 pi-sessao-7`
    Then the exit code is 1

  Scenario: done preserves the claim as history, reopen clears it
    When I run `mt create --vault <vault> "ciclo de vida do claim"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt claim --vault <vault> <id> pi-sessao-7`
    Then the exit code is 0
    When I run `mt done --vault <vault> <id>`
    Then the exit code is 0
    And the file "<vault>/issues/<id>.md" contains "claimed_by: pi-sessao-7"
    When I run `mt reopen --vault <vault> <id>`
    Then the exit code is 0
    And the file "<vault>/issues/<id>.md" does not contain "claimed_by:"
    When I run `mt ready --vault <vault>`
    Then the exit code is 0
    And stdout contains "ciclo de vida do claim"

  Scenario: a claimed done issue keeps the [claimed] marker as history under --all
    When I run `mt create --vault <vault> "historico de posse"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt claim --vault <vault> <id> pi-sessao-7`
    Then the exit code is 0
    When I run `mt done --vault <vault> <id>`
    Then the exit code is 0
    When I run `mt list --vault <vault>`
    Then the exit code is 0
    And stdout does not contain "historico de posse"
    When I run `mt list --vault <vault> --all`
    Then the exit code is 0
    And stdout contains "historico de posse [claimed]"

  Scenario: mt check rejects a blank claimed_by written by hand
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: claim com espaços
      status: open
      labels: []
      created_at: 2026-08-15T09:00
      claimed_by: "   "
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt check --vault <vault>`
    Then the exit code is 1
    And stderr contains "claimed_by"
