# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project follows Semantic Versioning.

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
