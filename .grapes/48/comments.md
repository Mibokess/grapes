### 2026-03-17T09:46
[STARTED] Adding GoReleaser + GitHub Actions release pipeline for cross-platform binary distribution.

### 2026-03-17T09:46
[VERIFY] `goreleaser check` passes. `go build` succeeds. All acceptance criteria met.

[DONE] Files changed:
- `main.go` — version default changed from `"0.1.7"` to `"dev"` (ldflags injects real version)
- `.goreleaser.yaml` — new, cross-platform build config
- `.github/workflows/release.yml` — new, triggers on `v*` tag push
- `.gitignore` — added `dist/`
