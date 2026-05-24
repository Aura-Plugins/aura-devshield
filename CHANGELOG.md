# Changelog

All notable changes to Aura DevShield are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [Unreleased] — Phase 1 distribution

### Added

- **`--version` flag** — prints the version string set at build time via `-ldflags "-X main.version=vX.Y.Z"`. Returns `dev` when built without ldflags.
- **`Makefile`** — `build`, `build-all` (cross-compile for darwin/arm64, darwin/amd64, linux/amd64, windows/amd64), `vet`, `install`, `checksums`, `clean` targets. Version is read from the current git tag automatically.
- **`.github/workflows/release.yml`** — GitHub Actions workflow that runs `go vet`, cross-compiles all targets, generates `checksums.txt`, and publishes a GitHub Release when a `v*` tag is pushed.
- **`install.sh`** — curl-pipe-bash install script. Detects OS and architecture, fetches the latest release from the GitHub API, downloads the binary, verifies the SHA256 checksum, and installs to `/usr/local/bin` (overridable via `INSTALL_DIR`).
- **`Formula/aura-devshield.rb`** — Homebrew formula template for publishing to a tap (`brew tap matias2018/tap && brew install aura-devshield`). Builds from source using the release tarball.

- **Update quarantine system** — new `vscode.update_in_quarantine` finding (Medium). Every extension version is stamped with a `first_seen` timestamp on first scan. Versions within the configurable quarantine window are flagged so updates are not blindly trusted. Default window: 7 days. Motivated by the wave of supply-chain attacks in May 2026.
- **`apply` subcommand** — previews and writes per-extension `extensions.autoUpdate` pins to VS Code `settings.json`. Dry-run by default; requires `--confirm` to write. Releases pins automatically when extensions clear the quarantine window.
- **`clean` subcommand** — finds and removes duplicate versions (keeping highest semver), malformed extensions (missing `name` or `publisher`), and orphaned directories. Dry-run by default; requires `--confirm` to delete.
- **`internal/state/` package** — persistent local state at `~/.aura-devshield/state.json`. Stores first-seen timestamps per extension version and tracks which extensions the tool has pinned.
- **`internal/config/` package** — loads `~/.aura-devshield/config.json`. Currently exposes `quarantine_days` (default `7`).
- **`internal/vscode/quarantine.go`** — `UpdateState`, `FindQuarantineFindings`, `QuarantinedExtensionIDs`.
- **`internal/vscode/settings.go`** — reads and writes VS Code `settings.json` with OS-aware path resolution (macOS, Linux, Windows). Uses per-extension `extensions.autoUpdate` map (VS Code 1.83+).
- **`internal/vscode/clean.go`** — `FindCleanTargets`, `Clean`, semver comparison via `compareVersions`.
- **`scanner.Scanner` interface** — foundation for future npm, GitHub Actions, Composer, and pip scanners.
- **`Metadata map[string]string` field on `Finding`** — carries structured per-finding context (e.g. `first_seen`, `days_remaining`, `quarantine_policy`).
- CLI subcommands: `scan`, `apply`, `clean`. Bare flags (e.g. `aura-devshield --json`) continue to work as `scan` for backwards compatibility.

### Changed

- `main.go` refactored from a flat script into a subcommand-based CLI using `flag.FlagSet` per command.
- `prepare()` helper centralises config/state loading, extension scanning, and state update across subcommands.
- `PrintFindings` now renders `Path` and `Metadata` fields when present.

### Fixed

- Severity casing inconsistency: `"low"` in `orphaned_findings.go` and `"Low"` in `symlink_findings.go` replaced with `scanner.SeverityLow`.

### Removed

- `internal/vscode/duplicates.go` — dead code, functionally identical to `multiversion.go`.

---

## [0.2.0]

Commits: `c48b5ed`, `52e5556`, `24ceac3`

### Added

- Orphaned directory detection: directories in the extensions folder without a `package.json` reported as `vscode.orphaned_extension_directory` (Low).
- Symlink detection: extension directories that are symbolic links reported as `vscode.symlinked_extension_directory` (Low).
- SHA256 fingerprints on every finding for deduplication.
- JSON output mode (`--json` flag) — suppresses all human output and emits valid JSON only.
- Deduplication infrastructure (`scanner.DeduplicateFindings`) — implemented but disabled by default to preserve instance-level findings.
- `internal/output/` package separating rendering from scanning logic (`terminal.go`, `json.go`).

### Changed

- Terminal output moved into dedicated `output.PrintExtensions` and `output.PrintFindings` functions.

---

## [0.1.0]

Commits: `e487505`, `29af517`, `e903f57`

### Added

- VS Code extension discovery via `~/.vscode/extensions`.
- `package.json` parsing into the `Extension` struct with `ID()` and `CanonicalID()` methods.
- Canonical ID normalisation (lowercase `publisher.name`) to handle publisher casing variations.
- Multi-version detection: two or more simultaneously installed versions of the same extension reported as `vscode.multiple_installed_versions` (Medium).
- Invalid metadata detection: empty and unresolved placeholder display names (e.g. `%displayName%`) reported as `vscode.invalid_metadata` (Low).
- Generic `Finding` struct with `ID`, `Fingerprint`, `Severity`, `Title`, `Description`, `Target`, `Path`.
- Severity levels: `Low`, `Medium`, `High`.
- `internal/scanner/` — reusable findings framework intended for all future scanner domains.
- Zero external dependencies — pure Go standard library.
