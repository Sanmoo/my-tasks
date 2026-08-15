Feature: Comment

  mt comment <id> <text> appends a comment to the Issue's Comments
  section: a ### timestamp heading, the text, and a stable
  <!-- comment: … --> anchor. The body is append-only — the existing
  content is preserved byte-for-byte. The append logic and the anchor
  token are decision-dense pure logic covered at Seam 2 (internal/issue);
  these scenarios cover the process: the compiled binary against a
  temporary Vault.

  Background:
    When I run `mt init --prefix pkm <vault>`
    Then the exit code is 0

  Scenario: comment appends a timestamped comment with a stable anchor
    When I run `mt create --vault <vault> "comprar material"`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt comment --vault <vault> <id> "Comprei metade da lista."`
    Then the exit code is 0
    And the file "<vault>/issues/<id>.md" matches "### [0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}"
    And the file "<vault>/issues/<id>.md" contains "Comprei metade da lista."
    And the file "<vault>/issues/<id>.md" matches "<!-- comment: [0-9a-f]{8} -->"

  Scenario: comment is append-only and preserves the existing body
    Given the file "<vault>/issues/pkm-0001.md" is written with:
      """
      ---
      title: t
      status: open
      labels: []
      created_at: 2026-08-15T09:30
      ---

      ## Description
      corpo original
      ## Notes
      ## Comments
      ### 2026-08-16T14:05
      Comprei metade da lista.
      <!-- comment: 4f2b9c1a -->
      """
    When I run `mt comment --vault <vault> pkm-0001 "Segunda anotação."`
    Then the exit code is 0
    And the file "<vault>/issues/pkm-0001.md" contains "corpo original"
    And the file "<vault>/issues/pkm-0001.md" contains "Comprei metade da lista."
    And the file "<vault>/issues/pkm-0001.md" contains "<!-- comment: 4f2b9c1a -->"
    And the file "<vault>/issues/pkm-0001.md" contains "Segunda anotação."
    And the file "<vault>/issues/pkm-0001.md" contains 2 occurrences of "<!-- comment: "

  Scenario: each comment gets its own stable anchor
    When I run `mt create --vault <vault> t`
    Then the exit code is 0
    And I remember the issue ID
    When I run `mt comment --vault <vault> <id> "primeiro"`
    Then the exit code is 0
    When I run `mt comment --vault <vault> <id> "segundo"`
    Then the exit code is 0
    And the file "<vault>/issues/<id>.md" contains "primeiro"
    And the file "<vault>/issues/<id>.md" contains "segundo"
    And the file "<vault>/issues/<id>.md" contains 2 occurrences of "<!-- comment: "

  Scenario: comment without a text argument is a usage error
    When I run `mt comment --vault <vault> pkm-0001`
    Then the exit code is 2
    And stderr contains "comment needs"
