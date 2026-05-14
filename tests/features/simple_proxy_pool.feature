Feature: Simple authenticated proxy support
  Other Go projects can use this module to parse simple authenticated proxy
  entries and expand IPv4 ranges without hardcoding proxy credentials.

  Scenario: Parse a simple proxy entry
    Given a proxy entry "40.27.182.2:3128:tgkoroke:tgkorfedwoksokfed"
    When the entry is parsed
    Then the proxy URL is "http://tgkoroke:tgkorfedwoksokfed@40.27.182.2:3128"

  Scenario: Expand the last IPv4 octet
    Given a proxy entry "40.27.182.2:3128:user:pass"
    When the last octet range is 1 through 3
    Then the generated proxy hosts are:
      | host        |
      | 40.27.182.1 |
      | 40.27.182.2 |
      | 40.27.182.3 |
