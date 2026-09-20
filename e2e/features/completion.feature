Feature: Shell completion of issue IDs

  `mt completion <shell>` (cobra-generated) drives the shell's <TAB>;
  every command that takes an issue key completes the vault's issue IDs
  through the same resolveVaultForKey the command itself uses — so
  completion inherits @bookmark > --vault > ID prefix > default and its
  ambiguity/fallback behavior. The hidden `mt __completeNoDesc` request
  prints one candidate per line, then the `:<directive>` line (4 =
  NoFileComp: the shell must not fall back to file completion). A
  resolution failure — an ambiguous prefix or no vault at all — prints
  zero candidates and only the directive: TAB stays silent and the
  command's own error appears on execution.

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
    And the file "<base>/vaults/dom/issues/dom-aaaa.md" is written with:
      """
      ---
      title: dom issue aaaa
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<base>/vaults/dom/issues/dom-abcd.md" is written with:
      """
      ---
      title: dom issue abcd
      status: open
      labels: []
      created_at: 2026-01-01T10:00
      ---

      ## Description
      ## Notes
      ## Comments
      """
    And the file "<base>/vaults/dom/issues/dom-abkg.md" is written with:
      """
      ---
      title: dom issue abkg
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

  Scenario: an ambiguous prefix lists the matching Issue IDs
    When I run `mt __completeNoDesc show dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    And stdout contains "dom-abkg"
    And stdout does not contain "dom-aaaa"

  Scenario: a unique prefix completes exactly one Issue ID
    When I run `mt __completeNoDesc show dom-abk`
    Then the exit code is 0
    And stdout contains "dom-abkg"
    And stdout does not contain "dom-abcd"
    And stdout does not contain "dom-aaaa"

  Scenario: the prefix filter is case-insensitive, the resolution is not
    When I run `mt __completeNoDesc show dom-AB`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    And stdout contains "dom-abkg"
    When I run `mt __completeNoDesc show DOM-AB`
    Then the exit code is 0
    And stdout does not contain "dom-abcd"
    And stdout matches ":4"

  Scenario: the vault prefix alone lists every Issue of the vault
    When I run `mt __completeNoDesc show dom-`
    Then the exit code is 0
    And stdout contains "dom-aaaa"
    And stdout contains "dom-abcd"
    And stdout contains "dom-abkg"
    And stdout does not contain "pkm-123"

  Scenario: an ambiguous key prefix yields zero candidates
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
    When I run `mt __completeNoDesc show dom-ab`
    Then the exit code is 0
    And stdout does not contain "dom-aaaa"
    And stdout does not contain "dom-abcd"
    And stdout matches ":4"

  Scenario: no resolvable vault yields zero candidates
    Given the file "<base>/config/mt/config.yaml" is written with:
      """
      """
    When I run `mt __completeNoDesc show dom-ab`
    Then the exit code is 0
    And stdout does not contain "dom-abcd"
    And stdout matches ":4"

  Scenario: dep add completes the subject and the blocker in the subject's vault
    When I run `mt __completeNoDesc dep add dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    And stdout contains "dom-abkg"
    When I run `mt __completeNoDesc dep add dom-abcd dom-`
    Then the exit code is 0
    And stdout contains "dom-aaaa"
    And stdout contains "dom-abcd"
    And stdout contains "dom-abkg"
    And stdout does not contain "pkm-123"
    When I run `mt __completeNoDesc dep add pkm-123 dom-`
    Then the exit code is 0
    And stdout does not contain "dom-aaaa"
    And stdout matches ":4"

  Scenario: dep rm completes the blocker in the subject's vault
    When I run `mt __completeNoDesc dep rm dom-abcd dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    And stdout contains "dom-abkg"
    And stdout does not contain "dom-aaaa"
    When I run `mt __completeNoDesc dep rm pkm-123 dom-ab`
    Then the exit code is 0
    And stdout does not contain "dom-abcd"
    And stdout matches ":4"

  Scenario: non-ID positions complete nothing
    When I run `mt __completeNoDesc status dom-abcd in`
    Then the exit code is 0
    And stdout matches ":4"
    When I run `mt __completeNoDesc rank dom-abcd 2`
    Then the exit code is 0
    And stdout matches ":4"
    When I run `mt __completeNoDesc comment dom-abcd a note`
    Then the exit code is 0
    And stdout matches ":4"
    When I run `mt __completeNoDesc defer dom-abcd +2d`
    Then the exit code is 0
    And stdout matches ":4"

  Scenario: every key command completes through the same function
    When I run `mt __completeNoDesc done dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    And stdout contains "dom-abkg"
    When I run `mt __completeNoDesc reopen dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    When I run `mt __completeNoDesc status dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    When I run `mt __completeNoDesc defer dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    When I run `mt __completeNoDesc undefer dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    When I run `mt __completeNoDesc top dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    When I run `mt __completeNoDesc bottom dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    When I run `mt __completeNoDesc unrank dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    When I run `mt __completeNoDesc rank dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    When I run `mt __completeNoDesc comment dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    When I run `mt __completeNoDesc edit dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    When I run `mt __completeNoDesc close dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"

  Scenario: --vault and the @bookmark are honored during completion
    When I run `mt __completeNoDesc show --vault <base>/vaults/dom dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    And stdout contains "dom-abkg"
    When I run `mt __completeNoDesc show @dom dom-ab`
    Then the exit code is 0
    And stdout contains "dom-abcd"
    And stdout contains "dom-abkg"
    When I run `mt __completeNoDesc show @pkm dom-ab`
    Then the exit code is 0
    And stdout does not contain "dom-abcd"
    And stdout matches ":4"