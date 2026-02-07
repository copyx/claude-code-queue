# CLAUDE.md

## Project
ccq (Claude Code Queue Manager) — Go CLI + Claude Code plugin in one repo.
FIFO queue-based auto-switcher for multiple Claude Code sessions via tmux.

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

## Architecture
- Hook-driven state machine, no daemon — tmux serializes all commands
- State stored in tmux window options with `@ccq_` prefix (`@ccq_state`, `@ccq_idle_since`, `@ccq_return_to`, `@ccq_auto_switch`)
- Plugin lives in `plugins/ccq/`, marketplace config in `.claude-plugin/marketplace.json`

## Gotchas
- `.gitignore` uses `/ccq` (root-only) to avoid matching `plugins/ccq/` directory
- tmux status bar conditionals: `#{?var,...}` tests non-empty, not value — use `#{?#{==:#{@var},on},...}` for string comparison
- `MarkIdle` must guard against repeated calls to preserve FIFO timestamp
