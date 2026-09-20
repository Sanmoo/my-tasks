Feature: Vault resolution by the issue ID prefix

  Commands that receive an issue key resolve the vault by the key's ID
  prefix — the part before the first '-' — when no @bookmark or --vault
  is given: `mt show dom-xyz` finds the vault whose prefix is dom
  without repeating @dom. The precedence is @bookmark > --vault > ID
  prefix > default bookmark: an explicit selection always wins, and a
  key without a prefix (or without a match) falls back to the default
  bookmark exactly as before, messages included. Errors exit 1, like
  every user error.

  Background:
    When I run `mt init --prefix pkm <base>/vaults/pkm`
    Then the exit code is 0
    When I run `mt init --prefix dom <base>/vaults/dom`
    Then the exit code is 0
    Given the file "<base>/config/mt/config.yaml" is written with:
      """
      default: pkm
      bookmarks:
        pkm: <base>/vaults/pkm
        dom: <base>/vaults/dom
      """
    And the file "<base>/vaults/dom/issues/dom-xyz.md" is written with:
      """
      ---
      title: dom issue
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<base>/vaults/pkm/issues/pkm-123.md" is written with:
      """
      ---
      title: pkm issue
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """

  Scenario: show finds the Issue in the vault its prefix announces
    When I run `mt show dom-xyz`
    Then the exit code is 0
    And stdout contains "dom issue"
    When I run `mt show pkm-123`
    Then the exit code is 0
    And stdout contains "pkm issue"

  Scenario: edit opens the file in the vault its prefix announces
    Given the fake editor writes
      """
      edited by the fake editor
      """
    When I run `mt edit dom-xyz`
    Then the exit code is 0
    And the file "<base>/vaults/dom/issues/dom-xyz.md" contains "edited by the fake editor"

  Scenario: status transitions operate in the vault the prefix announces
    When I run `mt status dom-xyz in_progress`
    Then the exit code is 0
    And stdout contains "dom-xyz is now in_progress: dom issue"
    When I run `mt done dom-xyz`
    Then the exit code is 0
    And stdout contains "dom-xyz is now done: dom issue"
    When I run `mt reopen dom-xyz`
    Then the exit code is 0
    And stdout contains "dom-xyz is now open: dom issue"
    And the file "<base>/vaults/pkm/issues/pkm-123.md" does not contain "status: done"

  Scenario: defer, comment and queue commands operate in the vault the prefix announces
    When I run `mt defer dom-xyz +3h`
    Then the exit code is 0
    And stdout contains "dom-xyz deferred until"
    When I run `mt comment dom-xyz a note`
    Then the exit code is 0
    And the file "<base>/vaults/dom/issues/dom-xyz.md" contains "a note"
    When I run `mt top dom-xyz`
    Then the exit code is 0
    And stdout contains "Updated 1 issue"
    And the file "<base>/vaults/dom/issues/dom-xyz.md" contains "rank: 1"
    And the file "<base>/vaults/pkm/issues/pkm-123.md" does not contain "rank:"
    When I run `mt unrank dom-xyz`
    Then the exit code is 0
    And stdout contains "Updated 1 issue"
    And the file "<base>/vaults/dom/issues/dom-xyz.md" does not contain "rank:"
    When I run `mt rank dom-xyz 1`
    Then the exit code is 0
    And stdout contains "Updated 1 issue"
    When I run `mt bottom dom-xyz`
    Then the exit code is 0

  Scenario: dep add records a blocker in the vault the IDs announce
    Given the file "<base>/vaults/dom/issues/dom-001.md" is written with:
      """
      ---
      title: other dom issue
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt dep add dom-xyz dom-001`
    Then the exit code is 0
    And stdout contains "dom-xyz is now blocked by dom-001"
    And the file "<base>/vaults/dom/issues/dom-xyz.md" contains "blocked_by: [dom-001]"

  Scenario: a missing Issue in the prefix-matched vault names the vault
    When I run `mt show dom-zzz`
    Then the exit code is 1
    And stderr contains "issue dom-zzz not found in @dom"
    When I run `mt done dom-zzz`
    Then the exit code is 1
    And stderr contains "issue dom-zzz not found in @dom"
    When I run `mt top dom-zzz`
    Then the exit code is 1
    And stderr contains "issue dom-zzz not found in @dom"

  Scenario: an explicit @bookmark wins and a matching prefix adds a hint
    When I run `mt show @pkm dom-xyz`
    Then the exit code is 1
    And stderr contains "issue dom-xyz not found in @pkm"
    And stderr contains "hint: the prefix"
    And stderr contains "suggests vault @dom"

  Scenario: an explicit --vault wins and a matching prefix adds a hint
    When I run `mt show --vault <base>/vaults/pkm dom-xyz`
    Then the exit code is 1
    And stderr contains "issue dom-xyz not found in <base>/vaults/pkm"
    And stderr contains "hint: the prefix"
    And stderr contains "suggests vault @dom"

  Scenario: an explicit selection matching only itself gets no hint
    When I run `mt show @dom dom-xyz`
    Then the exit code is 0
    And stdout contains "dom issue"

  Scenario: an ambiguous prefix lists the bookmarks and shows the disambiguation
    When I run `mt init --prefix dom <base>/vaults/a`
    Then the exit code is 0
    When I run `mt init --prefix dom <base>/vaults/b`
    Then the exit code is 0
    Given the file "<base>/config/mt/config.yaml" is written with:
      """
      default: pkm
      bookmarks:
        pkm: <base>/vaults/pkm
        a: <base>/vaults/a
        b: <base>/vaults/b
      """
    When I run `mt show dom-zzz`
    Then the exit code is 1
    And stderr contains "key dom-zzz is ambiguous: bookmarks @a and @b both use prefix"
    And stderr contains "pick one explicitly: mt show @a dom-zzz"
    When I run `mt top dom-zzz`
    Then the exit code is 1
    And stderr contains "key dom-zzz is ambiguous: bookmarks @a and @b both use prefix"
    And stderr contains "pick one explicitly: mt top @a dom-zzz"

  Scenario: two bookmarks for the same vault do not count as ambiguity
    Given the file "<base>/config/mt/config.yaml" is written with:
      """
      default: pkm
      bookmarks:
        pkm: <base>/vaults/pkm
        dom: <base>/vaults/dom
        alias: <base>/vaults/dom
      """
    When I run `mt show dom-zzz`
    Then the exit code is 1
    And stderr contains "issue dom-zzz not found in @alias"
    And stderr does not contain "ambiguous"

  Scenario: a key without a dash or without a match falls back to the default
    When I run `mt show legacy-id`
    Then the exit code is 1
    And stderr contains "issue legacy-id not found"
    When I run `mt done unknown-1`
    Then the exit code is 1
    And stderr contains "issue unknown-1 not found"
    And stderr does not contain "in @"

  Scenario: dep add rejects a blocker that belongs to another vault
    Given the file "<base>/vaults/dom/issues/dom-001.md" is written with:
      """
      ---
      title: other dom issue
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt dep add pkm-123 dom-001`
    Then the exit code is 1
    And stderr contains "blocker dom-001 belongs to @dom — blocked_by requires the same vault"
    And the file "<base>/vaults/pkm/issues/pkm-123.md" does not contain "blocked_by"

  Scenario: dep rm resolves only the subject, the blocker is a bare name
    Given the file "<base>/vaults/dom/issues/dom-xyz.md" is written with:
      """
      ---
      title: dom issue
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      blocked_by: [pkm-999]
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt dep rm dom-xyz pkm-999`
    Then the exit code is 0
    And stdout contains "dom-xyz is no longer blocked by pkm-999"
    And the file "<base>/vaults/dom/issues/dom-xyz.md" does not contain "blocked_by"

  Scenario: undefer with an ID clears that Issue in the vault the prefix announces
    Given the file "<base>/vaults/dom/issues/dom-xyz.md" is written with:
      """
      ---
      title: dom issue
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      deferred_until: 2999-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt undefer dom-xyz`
    Then the exit code is 0
    And stdout contains "Undeferred dom-xyz (was 2999-01-01T00:00)"
    And the file "<base>/vaults/dom/issues/dom-xyz.md" does not contain "deferred_until"

  Scenario: undefer without an ID sweeps the default vault
    Given the file "<base>/vaults/pkm/issues/pkm-123.md" is written with:
      """
      ---
      title: pkm issue
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      deferred_until: 2000-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<base>/vaults/dom/issues/dom-xyz.md" is written with:
      """
      ---
      title: dom issue
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      deferred_until: 2000-01-01T00:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    When I run `mt undefer`
    Then the exit code is 0
    And stdout contains "Undeferred pkm-123 (was 2000-01-01T00:00)"
    And stdout does not contain "dom-xyz"
    And the file "<base>/vaults/pkm/issues/pkm-123.md" does not contain "deferred_until"
    And the file "<base>/vaults/dom/issues/dom-xyz.md" contains "deferred_until: 2000-01-01T00:00"