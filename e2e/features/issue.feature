Feature: Issue create, show and edit

  mt create/q write an Issue file with the exact spec schema; show reads
  it back; edit opens it in $EDITOR. The frontmatter round-trip (stable
  order, optional fields only-when-set) is decision-dense pure logic,
  covered at Seam 2 (internal/issue); these scenarios cover the process:
  the compiled binary against a temporary Vault.

  Background:
    When I run `mt init --prefix pkm <vault>`
    Then the exit code is 0

  Scenario: create without a vault fails with instructions
    When I run `mt create comprar`
    Then the exit code is 1
    And stderr contains "@bookmark"
    And stderr contains "--vault"
    And stderr contains "default"

  Scenario: create writes a file with the exact schema
    When I run `mt create --vault <vault> "comprar material"`
    Then the exit code is 0
    And stdout contains "Created pkm-"
    And I remember the issue ID
    And the file "<vault>/issues/<id>.md" exists
    And the file "<vault>/issues/<id>.md" contains "title: comprar material"
    And the file "<vault>/issues/<id>.md" contains "status: open"
    And the file "<vault>/issues/<id>.md" contains "labels: []"
    And the file "<vault>/issues/<id>.md" matches "created_at: [0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}"
    And the file "<vault>/issues/<id>.md" contains "## Description"
    And the file "<vault>/issues/<id>.md" contains "## Notes"
    And the file "<vault>/issues/<id>.md" contains "## Comments"
    And the file "<vault>/issues/<id>.md" does not contain "id:"
    And the file "<vault>/issues/<id>.md" does not contain "updated_at:"

  Scenario: create accepts free labels
    When I run `mt create --vault <vault> --label compras --label familia comprar`
    Then the exit code is 0
    And I remember the issue ID
    And the file "<vault>/issues/<id>.md" contains "labels: [compras, familia]"

  Scenario: q prints only the ID
    When I run `mt q --vault <vault> "ideia rapida"`
    Then the exit code is 0
    And stdout matches "^pkm-[0-9a-z]{4}\n$"

  Scenario: two creates get distinct IDs
    When I run `mt q --vault <vault> um`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt q --vault <vault> dois`
    Then the exit code is 0
    And stdout does not contain "<id>"
    And the directory "<vault>/issues" contains 2 files

  Scenario: show displays a rendered view
    When I run `mt create --vault <vault> "comprar material"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt show --vault <vault> <id>`
    Then the exit code is 0
    And stdout matches "(?m)^○ <id> \. comprar material \[open\]$"
    And stdout matches "Created: [0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}"
    And stdout contains "## Description"
    And stdout contains "## Comments"
    And stdout does not contain "title:"
    And stdout does not contain "Rank:"

  Scenario: show renders optional fields only when set
    Given the file "<vault>/issues/pkm-0b4.md" is written with:
      """
      ---
      title: Comprar material
      status: in_progress
      labels: [compras, familia]
      created_at: 2026-06-26T18:00
      rank: 3
      deadline: 2026-08-30T00:00
      started_at: 2026-06-27T09:00
      blocked_by: [bjd-001, bjd-002]
      ---

      ## Description
      corpo
      """
    When I run `mt show --vault <vault> pkm-0b4`
    Then the exit code is 0
    And stdout matches "(?m)^◐ pkm-0b4 \. Comprar material \[in_progress\]$"
    And stdout contains "Labels: compras, familia"
    And stdout contains "Rank: 3"
    And stdout contains "Deadline: 2026-08-30 00:00"
    And stdout contains "Started: 2026-06-27 09:00"
    And stdout contains "Blocked by: bjd-001, bjd-002"
    And stdout does not contain "Completed:"
    And stdout does not contain "Deferred until:"

  Scenario: show renders ANSI codes when color is forced
    When I run `mt create --vault <vault> "comprar material"`
    Then the exit code is 0
    And I remember the issue ID
    Given the environment variable "CLICOLOR_FORCE" is "1"
    When I run `mt show --vault <vault> <id>`
    Then the exit code is 0
    And stdout contains "38;2;"
    And stdout contains "comprar material"
    And stdout does not contain "title:"

  Scenario: a malformed issue ID is a usage error
    When I run `mt show --vault <vault> a/b`
    Then the exit code is 2
    And stderr contains "invalid issue ID"

  Scenario: edit opens the editor and round-trips the file
    When I run `mt create --vault <vault> original`
    Then the exit code is 0
    And I remember the issue ID
    Given the fake editor writes
      """
      ---
      title: original
      status: open
      labels: []
      created_at: 2026-08-15T09:30
      ---

      ## Description
      corpo editado
      ## Notes
      ## Comments
      """
    When I run `mt edit --vault <vault> <id>`
    Then the exit code is 0
    And the file "<vault>/issues/<id>.md" contains "title: original"
    And the file "<vault>/issues/<id>.md" contains "corpo editado"
    When I run `mt show --vault <vault> <id>`
    Then the exit code is 0
    And stdout matches "(?m)^○ <id> \. original \[open\]$"
    And stdout contains "corpo editado"
