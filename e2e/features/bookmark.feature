Feature: Bookmark management

  mt bookmark add/list/rm manage bookmarks and the default bookmark in
  the auto-detected global config (XDG). A freshly added bookmark is
  addressable as @name by the next command; the process-level proof of
  that lands with the first vault-requiring command (create/show/edit,
  ticket 03), while the add → save → load → resolve round-trip is
  covered at Seam 2 (internal/vault).

  Scenario: add writes a bookmark to the global config
    When I run `mt bookmark add pkm <vault>`
    Then the exit code is 0
    And stdout contains "Bookmark pkm added"
    And the file "<base>/config/mt/config.yaml" exists
    And the file "<base>/config/mt/config.yaml" contains "pkm: <vault>"

  Scenario: add persists across commands and list shows it
    When I run `mt bookmark add pkm <vault>`
    And I run `mt bookmark list`
    Then the exit code is 0
    And stdout contains "pkm -> <vault>"

  Scenario: add rejects an invalid bookmark name
    When I run `mt bookmark add a/b <vault>`
    Then the exit code is 2
    And stderr contains "invalid bookmark name"

  Scenario: add rejects the @name form with a clear message
    When I run `mt bookmark add @pkm <vault>`
    Then the exit code is 2
    And stderr contains "take a bare name"

  Scenario: list shows bookmarks and marks the default
    Given the file "<base>/config/mt/config.yaml" is written with:
      """
      default: bjd
      bookmarks:
        bjd: ~/dev/bjd
        dom: ~/dev/dom
      """
    When I run `mt bookmark list`
    Then the exit code is 0
    And stdout contains "bjd -> ~/dev/bjd (default)"
    And stdout contains "dom -> ~/dev/dom"

  Scenario: rm removes one bookmark and leaves the rest intact
    Given the file "<base>/config/mt/config.yaml" is written with:
      """
      default: bjd
      bookmarks:
        bjd: ~/dev/bjd
        dom: ~/dev/dom
      """
    When I run `mt bookmark rm dom`
    Then the exit code is 0
    And stdout contains "Bookmark dom removed"
    And the file "<base>/config/mt/config.yaml" contains "bjd: ~/dev/bjd"
    And the file "<base>/config/mt/config.yaml" contains "default: bjd"
    And the file "<base>/config/mt/config.yaml" does not contain "dom:"

  Scenario: rm of the default bookmark clears the default
    Given the file "<base>/config/mt/config.yaml" is written with:
      """
      default: bjd
      bookmarks:
        bjd: ~/dev/bjd
        dom: ~/dev/dom
      """
    When I run `mt bookmark rm bjd`
    Then the exit code is 0
    And the file "<base>/config/mt/config.yaml" does not contain "default:"
    And the file "<base>/config/mt/config.yaml" contains "dom: ~/dev/dom"

  Scenario: rm of a missing bookmark fails with a clear message
    When I run `mt bookmark rm nope`
    Then the exit code is 1
    And stderr contains "bookmark @nope not found"
