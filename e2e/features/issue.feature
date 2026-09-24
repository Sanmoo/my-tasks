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

  Scenario: create --bottom puts the new Issue at the end of the queue
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
    When I run `mt create --vault <vault> --bottom "comprar material"`
    Then the exit code is 0
    And stdout matches "^Created pkm-[0-9a-z]{4} \(rank 3\)\n$"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 2"
    When I run `mt list --vault <vault>`
    Then stdout matches "^○ pkm-001  first\n○ pkm-002  second\n○ pkm-[0-9a-z]{4}  comprar material\n$"

  Scenario: create --top puts the new Issue first and shifts the queue
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
    When I run `mt create --vault <vault> --top "comprar material"`
    Then the exit code is 0
    And stdout matches "^Created pkm-[0-9a-z]{4} \(rank 1\)\n$"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 3"
    When I run `mt list --vault <vault>`
    Then stdout matches "^○ pkm-[0-9a-z]{4}  comprar material\n○ pkm-001  first\n○ pkm-002  second\n$"

  Scenario: create --bottom on an empty queue starts the queue at rank 1
    When I run `mt create --vault <vault> --bottom "comprar material"`
    Then the exit code is 0
    And stdout matches "^Created pkm-[0-9a-z]{4} \(rank 1\)\n$"
    When I run `mt list --vault <vault>`
    Then stdout matches "^○ pkm-[0-9a-z]{4}  comprar material\n$"

  Scenario: q --top places the new Issue first and prints only the ID
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
    When I run `mt q --vault <vault> --top "ideia rapida"`
    Then the exit code is 0
    And stdout matches "^pkm-[0-9a-z]{4}\n$"
    And I remember the issue ID
    And the file "<vault>/issues/<id>.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 2"

  Scenario: q --bottom places the new Issue last and prints only the ID
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
    When I run `mt q --vault <vault> --bottom "ideia rapida"`
    Then the exit code is 0
    And stdout matches "^pkm-[0-9a-z]{4}\n$"
    And I remember the issue ID
    And the file "<vault>/issues/<id>.md" contains "rank: 3"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 2"

  Scenario: create and q reject --top and --bottom together
    When I run `mt create --vault <vault> --top --bottom "x"`
    Then the exit code is 2
    And stderr contains "--top and --bottom"
    When I run `mt q --vault <vault> --top --bottom "x"`
    Then the exit code is 2
    And stderr contains "--top and --bottom"
    And the directory "<vault>/issues" contains 0 files

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

  Scenario: show one-line prints the same compact line as list
    Given the file "<vault>/issues/pkm-0b4.md" is written with:
      """
      ---
      title: Comprar material
      status: in_progress
      labels: []
      created_at: 2026-06-26T18:00
      ---

      ## Description
      corpo
      ## Notes
      ## Comments
      """
    When I run `mt show --vault <vault> --one-line pkm-0b4`
    Then the exit code is 0
    And stdout matches "^◐ pkm-0b4  Comprar material\\n$"
    And stdout does not contain "Created:"
    When I run `mt show --vault <vault> --oneline pkm-0b4`
    Then the exit code is 0
    And stdout matches "^◐ pkm-0b4  Comprar material\\n$"

  Scenario: show summary prints key metadata and the last comment
    Given the file "<vault>/issues/pkm-0b4.md" is written with:
      """
      ---
      title: Comprar material
      status: in_progress
      labels: []
      created_at: 2026-06-26T18:00
      deferred_until: 2999-08-23T00:00
      ---

      ## Description
      corpo
      ## Notes
      ## Comments
      ### 2026-08-16T14:05
      primeiro comentário
      <!-- comment: 4f2b9c1a -->
      ### 2026-08-16T15:05
      último comentário
      <!-- comment: 9a1b2c3d -->
      """
    When I run `mt show --vault <vault> --summary pkm-0b4`
    Then the exit code is 0
    And stdout contains "Title: Comprar material"
    And stdout contains "Key: pkm-0b4"
    And stdout contains "Deferred: yes"
    And stdout contains "Status: in_progress"
    And stdout contains "Last comment: último comentário"
    And stdout does not contain "## Description"

  Scenario: show rejects both compact output flags
    When I run `mt show --vault <vault> --one-line --summary pkm-0b4`
    Then the exit code is 2
    And stderr contains "mutually exclusive"

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
