# CLAUDE.md

## Project
ccq (Claude Code Queue Manager) — Go CLI + Claude Code plugin in one repo.
FIFO queue-based auto-switcher for multiple Claude Code sessions via tmux.
See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed design.

## Build & Test
- `make build` — build binary
- `make test` — run all tests
- `make install` — install to ~/.local/bin/ccq
- No external Go dependencies (no go.sum is normal)

## Release
- `git tag v<semver> && git push origin v<semver>` triggers GitHub Actions
- Builds darwin/linux × amd64/arm64 binaries, creates GitHub Release

## Conventions
- Commit messages: conventional commits (`feat:`, `fix:`, `ci:`, `docs:`, `test:`, `refactor:`)
- Versioning: semver
- Application language: English (user-facing messages, docs, comments)

## Workflow
- After implementation is complete, consolidate design content from `docs/plans/` into `CLAUDE.md` and `docs/ARCHITECTURE.md`, then delete the plan files

## Gotchas
- `.gitignore` uses `/ccq` (root-only) to avoid matching `plugins/ccq/` directory
- tmux status bar conditionals: `#{?var,...}` tests non-empty, not value — use `#{?#{==:#{@var},on},...}` for string comparison
- `MarkIdle` must guard against repeated calls to preserve FIFO timestamp
