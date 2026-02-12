---
name: release
description: Bump version in all manifests, commit, tag, and push a release
user_invocable: true
arguments:
  - name: version
    description: Semver version to release (e.g., 0.3.0)
    required: true
---

# Release

Automate the ccq release process: bump versions, commit, tag, and push.

## Steps

1. **Parse version argument**: Extract the version from `$ARGUMENTS`. Strip a leading `v` if present (e.g., `v0.3.0` → `0.3.0`).

2. **Validate semver format**: The version must match `MAJOR.MINOR.PATCH` (digits only, no pre-release suffix). If invalid, tell the user and stop.

3. **Check for clean working tree**: Run `git status --porcelain`. If there are uncommitted changes, warn the user and stop — they should commit or stash first.

4. **Run tests**: Execute `make test`. If any test fails, report the failure and stop.

5. **Update version in all 3 locations** using the Edit tool:
   - `.claude-plugin/marketplace.json` — `metadata.version`
   - `.claude-plugin/marketplace.json` — `plugins[0].version`
   - `plugins/ccq/.claude-plugin/plugin.json` — `version`

6. **Stage and commit**:
   ```
   git add .claude-plugin/marketplace.json plugins/ccq/.claude-plugin/plugin.json
   git commit -m "release: v{version}"
   ```

7. **Create git tag**: `git tag v{version}`

8. **Ask before pushing**: Tell the user the commit and tag are ready locally, then ask for confirmation before running `git push origin main --follow-tags`. Pushing the tag triggers the GitHub Actions release workflow.
