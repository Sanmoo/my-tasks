Feature: Place an Issue by pairwise comparison

  mt place discovers where an Issue belongs in the priority queue by asking,
  one Comparação at a time, which of two Issues is more prioritized. Each
  answer halves the interval of possible ranks — the optimal ⌈log2(N+1)⌉
  Comparações for a queue of N — and the rank is applied through the same
  quick-order plan as mt rank, which renormalizes 1..N. create and q run the
  same session right after writing the Issue, unless --no-place suppresses it.

  The session asks only when a human is at the terminal: with stdout on a pipe
  (a script, `ID=$(mt q ...)`, make audit) create/q keep the Issue in the
  Backlog without asking, and mt place refuses instead of no-op'ing. --answers
  drives a scripted session. The binary search itself is pure logic covered at
  Seam 2 (internal/priority); these scenarios cover the process: the compiled
  binary against a temporary Vault, with $EDITOR, stdout, stderr, exit code and
  the files on disk as the observable evidence. A pseudo-terminal (util-linux
  script(1)) reaches the interactive scenarios, and it merges stdout and stderr
  — so those assert on the merged text and on the files.

  Scenario: place without a vault fails with instructions
    When I run `mt place pkm-001 --answers a`
    Then the exit code is 1
    And stderr contains "@bookmark"
    And stderr contains "--vault"
    And stderr contains "default"

  Scenario: place without a terminal and without --answers is an error
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
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> pkm-002`
    Then the exit code is 1
    And stdout is empty
    And stderr contains "needs a terminal"
    And stderr contains "--answers"
    And the file "<vault>/issues/pkm-002.md" does not contain "rank:"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"

  Scenario: --answers walks the Comparações and applies the rank
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
      rank: 3
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-004.md" is written with:
      """
      ---
      title: new idea
      status: open
      labels: []
      created_at: 2026-01-04T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> --answers ab pkm-004`
    Then the exit code is 0
    And stdout contains "Placed pkm-004 at rank 2"
    And stderr contains "Which is more prioritized? (1/2)"
    And stderr contains "  a) ○ pkm-004  new idea"
    And stderr contains "  b) ○ pkm-002  second"
    And stderr contains "Which is more prioritized? (2/2)"
    And stderr contains "  b) ○ pkm-001  first"
    And the file "<vault>/issues/pkm-004.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 3"
    And the file "<vault>/issues/pkm-003.md" contains "rank: 4"

  Scenario: the session asks at most ⌈log2(N+1)⌉ Comparações
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
      rank: 3
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-004.md" is written with:
      """
      ---
      title: fourth
      status: open
      labels: []
      created_at: 2026-01-04T10:00
      rank: 4
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-005.md" is written with:
      """
      ---
      title: fifth
      status: open
      labels: []
      created_at: 2026-01-05T10:00
      rank: 5
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-006.md" is written with:
      """
      ---
      title: sixth
      status: open
      labels: []
      created_at: 2026-01-06T10:00
      rank: 6
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-007.md" is written with:
      """
      ---
      title: seventh
      status: open
      labels: []
      created_at: 2026-01-07T10:00
      rank: 7
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-008.md" is written with:
      """
      ---
      title: new idea
      status: open
      labels: []
      created_at: 2026-01-08T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> --answers aaa pkm-008`
    Then the exit code is 0
    And stdout contains "Placed pkm-008 at rank 1"
    And stderr contains "Which is more prioritized? (1/3)"
    And stderr contains "Which is more prioritized? (3/3)"
    And stderr does not contain "(4/"
    And the file "<vault>/issues/pkm-008.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 2"

  Scenario: answering i keeps the order the vault already has
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
      rank: 3
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<vault>/issues/pkm-004.md" is written with:
      """
      ---
      title: new idea
      status: open
      labels: []
      created_at: 2026-01-04T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> --answers i pkm-004`
    Then the exit code is 0
    And stdout contains "Placed pkm-004 at rank 3"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-004.md" contains "rank: 3"
    And the file "<vault>/issues/pkm-003.md" contains "rank: 4"

  Scenario: a queued Issue moves by the same session, comparing against others
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
      rank: 3
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> --answers aa pkm-003`
    Then the exit code is 0
    And stdout contains "Placed pkm-003 at rank 1"
    And stderr does not contain "b) ○ pkm-003"
    And the file "<vault>/issues/pkm-003.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 3"

  Scenario: an empty queue places at the first rank without asking
    Given the file "<vault>/issues/pkm-001.md" is written with:
      """
      ---
      title: new idea
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> --answers a pkm-001`
    Then the exit code is 0
    And stdout contains "Placed pkm-001 at rank 1"
    And stderr does not contain "Which is more prioritized?"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"

  Scenario: q cancels the session and nothing is applied
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
      title: new idea
      status: open
      labels: []
      created_at: 2026-01-03T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> --answers q pkm-003`
    Then the exit code is 1
    And stdout is empty
    And stderr contains "cancelled"
    And the file "<vault>/issues/pkm-003.md" does not contain "rank:"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 2"

  Scenario: a key that is not an answer is a usage error
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
      title: new idea
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> --answers x pkm-002`
    Then the exit code is 2
    And stdout is empty
    And stderr contains "--answers takes a, b, i or q"
    And the file "<vault>/issues/pkm-002.md" does not contain "rank:"

  Scenario: --answers that runs out is a usage error
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
      title: new idea
      status: open
      labels: []
      created_at: 2026-01-03T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> --answers a pkm-003`
    Then the exit code is 2
    And stderr contains "ran out"
    And the file "<vault>/issues/pkm-003.md" does not contain "rank:"

  Scenario: a done Issue cannot be placed
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
      title: finished
      status: done
      labels: []
      created_at: 2026-01-02T10:00
      rank: 2
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> --answers a pkm-002`
    Then the exit code is 1
    And stderr contains "is done and cannot be prioritized"
    And stderr does not contain "Which is more prioritized?"

  Scenario: an unknown Issue is rejected before the first Comparação
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
    When I run `mt place --vault <vault> --answers a pkm-999`
    Then the exit code is 1
    And stderr contains "unknown issue ID"
    And stderr does not contain "Which is more prioritized?"

  Scenario: a duplicate rank is refused before the session
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
      title: also first
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      rank: 1
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> --answers a pkm-002`
    Then the exit code is 1
    And stderr contains "duplicate rank: 1"
    And stderr contains "mt check --fix"
    And stderr does not contain "Which is more prioritized?"

  Scenario: create without a terminal keeps the Backlog and asks nothing
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
    When I run `mt init --prefix pkm <vault>`
    And I run `mt create --vault <vault> nova ideia`
    Then the exit code is 0
    And stdout contains "Created pkm-"
    And stdout does not contain "Placed"
    And stderr does not contain "Which is more prioritized?"
    And the directory "<vault>/issues" contains 3 files
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 2"

  Scenario: q keeps its ID-only output while --answers places the Issue
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
    When I run `mt init --prefix pkm <vault>`
    And I run `mt q --vault <vault> --answers ab quieta`
    Then the exit code is 0
    And stdout does not contain "Which is more prioritized?"
    And stderr contains "Which is more prioritized? (1/2)"
    And I remember the issue ID
    And the file "<vault>/issues/<id>.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 3"

  Scenario: create --answers places the new Issue and keeps stdout unchanged
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
    When I run `mt init --prefix pkm <vault>`
    And I run `mt create --vault <vault> --answers aa nova ideia`
    Then the exit code is 0
    And stdout contains "Created pkm-"
    And stdout does not contain "rank"
    And stderr contains "Which is more prioritized? (1/2)"
    And stderr contains "Placed pkm-"
    And the directory "<vault>/issues" contains 3 files
    And the file "<vault>/issues/pkm-001.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 3"

  Scenario: contradictory create placement flags are usage errors
    When I run `mt create --vault <vault> --no-place --top x`
    Then the exit code is 2
    And stderr contains "--no-place"
    When I run `mt create --vault <vault> --no-place --answers a x`
    Then the exit code is 2
    And stderr contains "--answers"
    When I run `mt q --vault <vault> --answers a --bottom x`
    Then the exit code is 2
    And stderr contains "--answers"

  Scenario: create asks at a terminal and places the new Issue
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
    When I run `mt init --prefix pkm <vault>`
    And I run `mt create --vault <vault> nova ideia` at a terminal, typing:
      """
      a
      a
      """
    Then the exit code is 0
    And stdout contains "Created pkm-"
    And stdout contains "Which is more prioritized? (1/2)"
    And stdout contains "Placed pkm-"
    And the directory "<vault>/issues" contains 3 files
    And the file "<vault>/issues/pkm-001.md" contains "rank: 2"
    And the file "<vault>/issues/pkm-002.md" contains "rank: 3"

  Scenario: create --no-place asks nothing even at a terminal
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
    When I run `mt init --prefix pkm <vault>`
    And I run `mt create --vault <vault> --no-place nova ideia` at a terminal, typing:
      """
      a
      """    Then the exit code is 0
    And stdout contains "Created pkm-"
    And stdout does not contain "Which is more prioritized?"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
    And the directory "<vault>/issues" contains 2 files

  Scenario: an invalid key at the terminal is re-asked
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
      title: new idea
      status: open
      labels: []
      created_at: 2026-01-03T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> pkm-003` at a terminal, typing:
      """
      z
      a
      b
      """
    Then the exit code is 0
    And stdout contains "Answer a, b, i or q."
    And stdout contains "Placed pkm-003 at rank 2"
    And the file "<vault>/issues/pkm-003.md" contains "rank: 2"

  Scenario: the end of input cancels a terminal session
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
      title: new idea
      status: open
      labels: []
      created_at: 2026-01-02T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt place --vault <vault> pkm-002` at a terminal, typing:
      """
      z
      """
    Then the exit code is 1
    And stdout contains "cancelled"
    And stdout contains "the queue is unchanged"
    And the file "<vault>/issues/pkm-002.md" does not contain "rank:"
    And the file "<vault>/issues/pkm-001.md" contains "rank: 1"
