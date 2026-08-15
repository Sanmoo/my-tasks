Feature: CLI basics

  Smoke of the mt skeleton: the binary compiles, help works, and the exit
  code convention (0 success / 1 user error / 2 usage error) holds.
  CLI scenarios run the compiled binary — no function is called in-process —
  and the two harness scenarios verify the seams every scenario gets.

  Scenario: help prints usage and exits zero
    When I run `mt --help`
    Then the exit code is 0
    And stdout contains "Usage:"
    And stdout contains "Exit codes"

  Scenario: an unknown command is a usage error
    When I run `mt frobnicate`
    Then the exit code is 2
    And stderr contains "unknown command"

  Scenario: an unknown flag is a usage error
    When I run `mt --frobnicate`
    Then the exit code is 2
    And stderr contains "unknown flag"

  Scenario: the harness provides a temporary vault for every scenario
    Then a temporary vault exists
    And the vault contains an issues directory

  Scenario: the harness provides a fake editor for every scenario
    Then a fake editor is available
