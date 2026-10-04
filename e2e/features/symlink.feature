Feature: A symlink in issues/ is not an Issue

  The only behavior change of the Issue file store: a symlink inside
  issues/ stops being treated as an Issue. mt list skips it — a stray
  link does not bring the Vault down — and mt show refuses it instead of
  following the link and reading outside the Vault. The full file policy
  is covered black-box at Seam 2 (internal/issuefiles); these scenarios
  fix the visible contract.

  Background:
    When I run `mt init --prefix pkm <vault>`
    Then the exit code is 0

  Scenario: list skips a symlink in issues/
    Given the file "<vault>/issues/pkm-real.md" is written with:
      """
      ---
      title: a real issue
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the symbolic link "<vault>/issues/pkm-link.md" points to "<base>/outside.md"
    And the file "<base>/outside.md" is written with:
      """
      ---
      title: outside the vault
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt list --vault <vault>`
    Then the exit code is 0
    And stdout contains "pkm-real"
    And stdout contains "a real issue"
    And stdout does not contain "pkm-link"
    And stdout does not contain "outside the vault"

  Scenario: show refuses a symlink without reading its target
    Given the symbolic link "<vault>/issues/pkm-link.md" points to "<base>/outside.md"
    And the file "<base>/outside.md" is written with:
      """
      ---
      title: outside the vault
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt show pkm-link --vault <vault>`
    Then the exit code is 1
    And stderr contains "pkm-link"
    And stderr contains "not a regular file"
    And stdout does not contain "outside the vault"
