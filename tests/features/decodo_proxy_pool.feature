Feature: Decodo proxy pool for Go clients
  As a Go application using Decodo residential proxies
  I want a package that builds Decodo proxy settings and manages sticky sessions
  So that I can reuse the same package with httpcloak and net/http

  Scenario: Build a rotating proxy URL
    Given valid Decodo user credentials
    When I build a rotating proxy without sticky session settings
    Then I get a proxy URL targeting gate.decodo.com:7000
    And the generated username starts with "user-"
    And the generated username does not include a session parameter

  Scenario: Build a sticky proxy URL with targeting
    Given valid Decodo user credentials
    And a US targeting with city "new_york"
    When I build a sticky proxy with session id "task-1" and duration 30 minutes
    Then the generated username includes country, city, session, and sessionduration parameters

  Scenario: Reuse a sticky session by business key
    Given a pool configured for sticky sessions
    When I request a proxy for business key "account-1" twice before expiry
    Then I receive the same session id both times

  Scenario: Rotate a sticky session after reported failures
    Given a pool configured for sticky sessions
    And failure threshold 2
    When I report two failures for business key "account-1"
    Then the next proxy for business key "account-1" uses a new session id

  Scenario: Reject invalid proxy username input early
    Given a raw Decodo proxy username containing spaces, escapes, or other invalid symbols
    When I create auth credentials with NewAuth
    Then I get a local validation error instead of reaching proxy authentication

  Scenario: Accept safe characters in raw proxy username input
    Given a raw Decodo proxy username using only letters, digits, dot, underscore, and hyphen
    When I create auth credentials with NewAuth
    Then the username is accepted unchanged
