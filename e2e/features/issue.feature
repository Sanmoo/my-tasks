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

  Scenario: show displays frontmatter and body
    When I run `mt create --vault <vault> "comprar material"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt show --vault <vault> <id>`
    Then the exit code is 0
    And stdout contains "title: comprar material"
    And stdout contains "status: open"
    And stdout contains "## Description"
    And stdout contains "## Comments"

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
    And stdout contains "title: original"
    And stdout contains "corpo editado"
