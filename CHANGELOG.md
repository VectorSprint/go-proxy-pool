# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project follows Semantic Versioning.

## [Unreleased]

### Added

- Nothing yet.

## [0.4.0] - 2026-05-14

### Added

- `pkg/simpleproxy` for parsing plain `ip:port:username:password` proxy entries
- `simpleproxy.IPv4Range` to expand the last IPv4 octet dynamically without hardcoding full proxy lists
- README and BDD documentation for simple authenticated proxy usage

## [0.3.0] - 2026-04-19

### Added

- Repository-level `.golangci.yml` and `Taskfile.yml` for local `test`, `lint`, and `check` commands

## [0.2.0] - 2026-03-25

### Added

- Explicit `EndpointSpec` and `PortRange` support for Decodo dedicated endpoints
- Automatic rotating-port and sticky-port selection from dedicated endpoint metadata
- Sticky pool port allocation within configured sticky port ranges

## [0.1.0] - 2026-03-25

### Added

- Initial `decodo` package with typed `Auth`, `Config`, `Targeting`, and `Session` models
- Decodo proxy username and proxy URL generation for user:pass backconnect
- Sticky-session pool with keyed lease reuse, manual rotation, expiry cleanup, and failure-driven rotation
- Adapter helpers for `httpcloak` and Go `net/http`
- Unit tests, executable examples, and BDD feature coverage
- Go doc comments, MIT `LICENSE`, `.gitignore`, and GitHub Actions Go test workflow

### Notes

- Current module path: `github.com/VectorSprint/go-proxy-pool`
- Current focus: Decodo residential proxy integration for Go applications
