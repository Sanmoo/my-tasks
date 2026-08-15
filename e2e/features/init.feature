Feature: Vault init

  mt init creates a usable Vault: the issues/ directory plus the vault
  config (mt.yaml) with the ID prefix and status list. Vault resolution
  (@bookmark > --vault > default) is decision-dense pure logic, covered
  at Seam 2 (internal/vault); its process-level scenarios land with the
  first vault-requiring command (create/show/edit, ticket 03).

  Scenario: init creates a usable vault with the default config
    When I run `mt init <base>/fresh`
    Then the exit code is 0
    And stdout contains "Vault ready at"
    And the directory "<base>/fresh/issues" exists
    And the file "<base>/fresh/mt.yaml" exists
    And the file "<base>/fresh/mt.yaml" contains "prefix: fresh"
    And the file "<base>/fresh/mt.yaml" contains "status: [open, in_progress, done]"

  Scenario: init accepts a custom ID prefix and status list
    When I run `mt init --prefix PKM --status open --status in_progress --status blocked <base>/pkm`
    Then the exit code is 0
    And the file "<base>/pkm/mt.yaml" contains "prefix: PKM"
    And the file "<base>/pkm/mt.yaml" contains "status: [open, in_progress, blocked]"

  Scenario: init on an existing vault is refused
    When I run `mt init <vault>`
    Then the exit code is 0
    And the file "<vault>/mt.yaml" exists
    When I run `mt init <vault>`
    Then the exit code is 1
    And stderr contains "already exists"

  Scenario: init does not take a bookmark
    When I run `mt init @pkm <base>/fresh`
    Then the exit code is 2
    And stderr contains "@bookmark"

  Scenario: init does not take --vault
    When I run `mt init --vault <base>/x`
    Then the exit code is 2
    And stderr contains "--vault"

  Scenario: bare init derives the prefix from the current directory
    When the working directory is "<base>/cwd"
    When I run `mt init`
    Then the exit code is 0
    And the file "<base>/cwd/mt.yaml" contains "prefix: cwd"
    And the directory "<base>/cwd/issues" exists

  Scenario: init without a derivable prefix asks for one
    When I run `mt init <base>/@!`
    Then the exit code is 1
    And stderr contains "--prefix"
