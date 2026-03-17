## Goal
Add GoReleaser configuration and a GitHub Actions workflow so that pushing a version tag (e.g. `v0.2.0`) automatically builds cross-platform binaries and publishes a GitHub Release.

## Context
- Entry point: `main.go` at repo root
- Version previously hardcoded: `var version = "0.1.7"` in `main.go:14`
- Module path: `github.com/Mibokess/grapes`
- Go version: 1.24.2
- No existing CI/CD or release automation
- Users currently must clone the repo and run `go build`

## Acceptance Criteria
- [x] `.goreleaser.yaml` exists with cross-platform build config (linux/darwin/windows, amd64/arm64)
- [x] Version injected via ldflags from git tag (no more hardcoded version)
- [x] `.github/workflows/release.yml` triggers on `v*` tag push and runs GoReleaser
- [x] `.gitignore` updated with `dist/` (GoReleaser output directory)
- [x] `main.go` version variable set via ldflags, with `dev` as default for local builds

## Verify
```bash
go run github.com/goreleaser/goreleaser/v2@latest check
```

## Pass Criteria
GoReleaser check passes without errors.
