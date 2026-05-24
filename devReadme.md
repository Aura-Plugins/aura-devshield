# Aura DevShield — Developer Reference

Architecture, design decisions, and instructions for extending the tool.

---

## Repository layout

```
aura-devshield/
├── cmd/
│   └── aura-devshield/
│       └── main.go              # CLI entry point — subcommand routing and orchestration
├── internal/
│   ├── config/
│   │   └── config.go            # Config file loading (~/.aura-devshield/config.json)
│   ├── output/
│   │   ├── json.go              # JSON output formatter
│   │   └── terminal.go          # Human-readable output formatter
│   ├── scanner/
│   │   ├── deduplicate.go       # Dedup by fingerprint (infrastructure, disabled by default)
│   │   ├── finding.go           # Finding struct + Severity enum
│   │   ├── fingerprint.go       # SHA256 fingerprint generation
│   │   └── scanner.go           # Scanner interface
│   ├── state/
│   │   └── state.go             # Persistent state (~/.aura-devshield/state.json)
│   └── vscode/
│       ├── clean.go             # CleanTarget detection + os.RemoveAll execution
│       ├── discover.go          # Orphaned directory detection
│       ├── extensions.go        # ListExtensions (os.ReadDir wrapper)
│       ├── findings.go          # FindMultiVersionFindings
│       ├── metadata_findings.go # FindInvalidMetadataFindings
│       ├── multiversion.go      # FindMultiVersionExtensions (grouping logic)
│       ├── orphaned_findings.go # FindOrphanedDirectoryFindings
│       ├── packagejson.go       # ReadPackageJSON
│       ├── paths.go             # ExtensionsDir resolution
│       ├── quarantine.go        # Quarantine findings + state update
│       ├── scanner.go           # ScanExtensions orchestrator
│       ├── settings.go          # VS Code settings.json read/write
│       ├── symlink_findings.go  # FindSymlinkedExtensionDirectoryFindings
│       ├── symlinks.go          # FindSymlinkedExtensionDirs
│       └── types.go             # Extension struct
├── CHANGELOG.md
├── devReadme.md                 # this file
├── go.mod
└── README.md
```

---

## Core data types

### `scanner.Finding`

The universal output unit. Every scanner, regardless of domain, produces `[]scanner.Finding`.

```go
type Finding struct {
    ID          string            // namespaced: "vscode.multiple_installed_versions"
    Fingerprint string            // SHA256 for deduplication
    Severity    Severity          // SeverityLow | SeverityMedium | SeverityHigh
    Title       string            // short human label
    Description string            // full explanation
    Target      string            // what the finding points at (extension ID, path, etc.)
    Path        string            // filesystem path, if applicable
    Metadata    map[string]string // structured extra context (e.g. days_remaining, first_seen)
}
```

**ID naming convention:** `<scanner_domain>.<finding_type>` in snake_case.
Examples: `vscode.update_in_quarantine`, `npm.unpinned_dependency`, `github_actions.unpinned_action`.

**Always use severity constants** — never string literals. The type is `Severity string` so the compiler won't catch a literal like `"low"` (wrong case). Use `scanner.SeverityLow`, `scanner.SeverityMedium`, `scanner.SeverityHigh`.

---

### `scanner.Scanner` (interface)

```go
type Scanner interface {
    Name() string
    Scan() ([]Finding, error)
}
```

Every scanner must implement this. `Name()` returns the domain identifier (`"vscode"`, `"npm"`, etc.). The interface exists today so `main.go` can eventually run all scanners in a loop rather than calling each one explicitly.

---

### `state.State`

Persistent cross-run state at `~/.aura-devshield/state.json`.

```go
type State struct {
    VSCodeExtensions map[string]map[string]time.Time `json:"vscode_extensions"` // canonicalID → version → first_seen
    VSCodePinned     map[string]bool                 `json:"vscode_pinned,omitempty"` // extensions this tool has pinned
}
```

The `RecordVSCodeExtension` method is a no-op for existing entries — the first-seen timestamp is immutable once written. Future scanners needing temporal tracking should add their own top-level key to this struct and corresponding methods, following the same pattern.

---

## Data flow

```
main() — subcommand routing
  │
  └── prepare()
        ├── config.Load()           read ~/.aura-devshield/config.json (or defaults)
        ├── state.Load()            read ~/.aura-devshield/state.json (or empty state)
        ├── vscode.ExtensionsDir()  resolve ~/.vscode/extensions
        ├── vscode.ScanExtensions() ReadPackageJSON per directory → []*Extension
        └── vscode.UpdateState()    stamp first_seen for new versions (no-op for existing)

  ├── scan    → finding generators → output.PrintFindings / PrintFindingsJSON
  │             then state.Save()
  │
  ├── apply   → QuarantinedExtensionIDs → toPin / toRelease
  │             → PreviewQuarantine or ApplyQuarantine
  │             → state.Save() (updates VSCodePinned)
  │
  └── clean   → FindCleanTargets → PrintCleanTargets
                → vscode.Clean (if --confirm)
```

---

## Adding a new scanner

This is the intended extension path for npm, GitHub Actions, Composer, and pip.

### Step 1 — Create the package

```
internal/npm/
├── scanner.go     # implements scanner.Scanner
├── types.go       # Package struct (analogous to vscode.Extension)
├── lockfile.go    # parse package-lock.json / yarn.lock
└── findings.go    # finding generators
```

### Step 2 — Implement `scanner.Scanner`

```go
package npm

import "github.com/matias2018/aura-devshield/internal/scanner"

type Scanner struct {
    WorkDir string // directory containing package.json
}

func (s *Scanner) Name() string { return "npm" }

func (s *Scanner) Scan() ([]scanner.Finding, error) {
    // 1. Parse package-lock.json / yarn.lock
    // 2. Run finding generators
    // 3. Return []scanner.Finding
    return nil, nil
}
```

### Step 3 — Define finding IDs

Use the `npm.` prefix:

```
npm.unpinned_dependency          — dep uses a range rather than an exact version
npm.recently_published_package   — version published within quarantine window
npm.lock_mismatch                — package-lock.json out of sync with package.json
```

### Step 4 — Add state support if the scanner needs temporal tracking

Extend `state.State` in `internal/state/state.go`:

```go
type State struct {
    VSCodeExtensions map[string]map[string]time.Time `json:"vscode_extensions"`
    VSCodePinned     map[string]bool                 `json:"vscode_pinned,omitempty"`
    NpmPackages      map[string]map[string]time.Time `json:"npm_packages,omitempty"` // ← add
}
```

Add `RecordNpmPackage` and `NpmFirstSeen` methods following the same pattern as the vscode equivalents. Initialise the map in both `Load` (when nil after unmarshal) and the empty-state return.

### Step 5 — Wire into `main.go`

Add findings to `runScan`:

```go
npmScanner := &npm.Scanner{WorkDir: "."}
npmFindings, err := npmScanner.Scan()
if err != nil && !*jsonOutput {
    fmt.Fprintf(os.Stderr, "Warning: npm scan: %v\n", err)
} else {
    findings = append(findings, npmFindings...)
}
```

Once there are three or more scanners, replace the explicit calls with a loop:

```go
scanners := []scanner.Scanner{
    &vscode.Scanner{ExtensionsDir: extensionsDir},
    &npm.Scanner{WorkDir: "."},
    &githubactions.Scanner{WorkDir: "."},
}
for _, s := range scanners {
    f, err := s.Scan()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Warning: %s scan: %v\n", s.Name(), err)
        continue
    }
    findings = append(findings, f...)
}
```

### Step 6 — Add output support

If the new scanner needs a `clean`-style action (e.g. npm's equivalent would be `npm dedupe`), add it to the relevant subcommand in `main.go` and add print functions to `internal/output/terminal.go`. Follow the existing `PrintCleanTargets` / `PrintQuarantineResults` pattern.

---

## Design decisions and rationale

### Zero external dependencies

The entire tool uses only the Go standard library (`os`, `path/filepath`, `encoding/json`, `crypto/sha256`, `encoding/hex`, `flag`, `fmt`, `runtime`, `strings`, `strconv`, `sort`, `time`).

Reasons:
- Single static binary — trivially distributed, no install step.
- No supply-chain irony: a supply-chain visibility tool should not itself pull untrusted packages.
- No `go.sum` audit burden for contributors.

**Rule:** do not add an external dependency without a compelling reason and explicit discussion. The stdlib can handle JSON, HTTP, file I/O, and crypto — which covers everything planned.

### Local state over marketplace API

The quarantine system tracks first-seen timestamps locally rather than querying the VS Code Marketplace API for publish dates. This preserves the no-outbound-connections philosophy.

The tradeoff: "first seen" means "first time the user ran scan", not "when the version was published". In practice, a day or two of lag before a user runs the tool is acceptable within a 7-day window.

If marketplace API support is added in the future, make it opt-in via a `--check-remote` flag and implement it in a dedicated `internal/marketplace/` package — not inline in the scanner.

### Dry-run by default for destructive commands

`apply` and `clean` require explicit `--confirm` to write or delete anything. The tool's philosophy is visibility before enforcement. Users should always see what will happen before it happens. Never make a destructive command the default.

### Per-extension `extensions.autoUpdate` map

VS Code 1.83+ supports `"extensions.autoUpdate": { "publisher.name": false }` in addition to the global boolean. The `apply` command uses this per-extension form so it only blocks auto-update for quarantined extensions, not globally.

If a user has already set `extensions.autoUpdate: false` globally, the tool leaves it untouched and still tracks its own quarantine state via `VSCodePinned`. The `pinnedBySettings` function in `settings.go` only reads the map form; it ignores global booleans.

### Finding fingerprints

Every finding carries a SHA256 fingerprint computed from its key fields. `scanner.DeduplicateFindings` is implemented but disabled in `main.go` by default, preserving instance-level findings (e.g. each duplicate version gets its own finding). A future `--dedupe` flag can expose deduplication. Fingerprints are also useful for downstream tooling: CI gates, audit trails, SARIF export.

### Import dependency rules

The import graph must stay acyclic and layered:

```
main
  → output   → vscode, scanner
  → vscode   → scanner, state
  → config   (no internal imports)
  → state    (no internal imports)
  → scanner  (no internal imports)
```

`output` may import `vscode` (it already does for `PrintExtensions`). `vscode` must not import `output`. `scanner`, `state`, and `config` must not import any other internal package.

---

## Development workflow

### Build

```bash
go build ./...
```

### Vet (run before every commit)

```bash
go vet ./...
```

### Run subcommands

```bash
go run ./cmd/aura-devshield scan
go run ./cmd/aura-devshield scan --json
go run ./cmd/aura-devshield apply
go run ./cmd/aura-devshield apply --confirm
go run ./cmd/aura-devshield clean
go run ./cmd/aura-devshield clean --confirm
```

### Reset first-seen timestamps during development

```bash
rm ~/.aura-devshield/state.json
```

### Back up VS Code settings before testing `apply --confirm`

```bash
cp ~/Library/Application\ Support/Code/User/settings.json \
   ~/Library/Application\ Support/Code/User/settings.json.bak
```

### Check the state file

```bash
cat ~/.aura-devshield/state.json | python3 -m json.tool | head -40
```

---

## Planned scanner roadmap

| Package | Target files | Key finding types |
|---|---|---|
| `internal/npm/` | `package-lock.json`, `yarn.lock`, `node_modules/` | unpinned deps, recently published versions |
| `internal/githubactions/` | `.github/workflows/*.yml` | unpinned `@latest` action refs, actions from unverified publishers |
| `internal/composer/` | `composer.lock`, `composer.json` | recently updated packages, missing lock file |
| `internal/pip/` | `requirements.txt`, `pyproject.toml`, `pip.lock` | recently updated packages, unpinned versions |

Each scanner lives in its own package under `internal/`, implements `scanner.Scanner`, and follows the file layout established by `internal/vscode/`.

---

## What Aura DevShield is not

This context matters for scoping features and rejecting out-of-scope requests:

- Not an antivirus or signature scanner
- Not an EDR or kernel monitor
- Not a behavioral analysis engine
- Not a replacement for runtime security tooling
- Not a corporate SOC integration

It is a **local development environment visibility tool**. The goal is to surface information developers need to make informed decisions — not to make those decisions for them, and not to silently remediate anything.

## Future
Phase 1 (now): Ship the CLI as a binary via GitHub Releases + a Homebrew formula. That's your primary developer audience.

Phase 2: Add a macOS menu bar status indicator as a lightweight companion — not a full UI. It runs aura-devshield scan --json periodically and shows a badge if findings exist. This gives ambient awareness without replacing the CLI. Go has good support for this via the systray package (the one meaningful external dep worth considering).

Phase 3: Only build a full UI if you're targeting non-developers (security teams, compliance, management). At that point a web report (scan --output report.html) is lighter than a full SPA and keeps the zero-server-process philosophy.

The CLI is the core. Any UI is a presentation layer on top of it.

## Built
Phase 1 is done. Here's what was built:

Deliverable	What it does
Makefile	make build (current platform), make build-all (4 targets), make install, make vet, make checksums
.github/workflows/release.yml	Triggers on v* tag push → vets → cross-compiles → generates checksums.txt → publishes GitHub Release
install.sh	Detects OS/arch, fetches latest tag from GitHub API, downloads binary, verifies SHA256, installs to /usr/local/bin
Formula/aura-devshield.rb	Homebrew formula (builds from source tarball). Needs a homebrew-tap repo to publish. SHA256 placeholder to fill after first release.
--version flag	aura-devshield --version prints the version injected at build time via -ldflags
To ship the first release:

Tag the repo: git tag v0.1.0 && git push origin v0.1.0
GitHub Actions builds and publishes automatically
Update Formula/aura-devshield.rb with the SHA256 of the v0.1.0 tarball and push the formula to your homebrew-tap repo
