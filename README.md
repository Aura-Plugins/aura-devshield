# Aura DevShield

A local-first, privacy-focused security visibility tool for developer environments.

Aura DevShield inspects your development tooling — starting with VS Code extensions — and surfaces supply-chain risks before they become incidents. No telemetry. No cloud. No account required.

---

## The problem

Supply-chain attacks targeting developer tools have increased sharply. A compromised VS Code extension, npm package, or GitHub Action can silently exfiltrate secrets, modify builds, or persist across reinstalls. Most developers have no visibility into what is installed, what has changed, or when it changed.

Aura DevShield gives you that visibility, and gives newly-released updates a cooling-off period before they are trusted.

---

## Install

### Script (macOS and Linux — recommended)

```bash
curl -sSfL https://raw.githubusercontent.com/matias2018/aura-devshield/main/install.sh | bash
```

Installs to `/usr/local/bin`. Override the directory with `INSTALL_DIR`:

```bash
INSTALL_DIR=~/.local/bin bash <(curl -sSfL https://raw.githubusercontent.com/matias2018/aura-devshield/main/install.sh)
```

### Homebrew (macOS and Linux)

```bash
brew tap matias2018/tap
brew install aura-devshield
```

### Download binary directly

Pre-built binaries for every release are on the [GitHub Releases](https://github.com/matias2018/aura-devshield/releases) page. Download the binary for your platform, make it executable, and move it to your PATH.

| Platform | Binary |
|---|---|
| macOS Apple Silicon | `aura-devshield-darwin-arm64` |
| macOS Intel | `aura-devshield-darwin-amd64` |
| Linux x86-64 | `aura-devshield-linux-amd64` |
| Windows x86-64 | `aura-devshield-windows-amd64.exe` |

Each release includes a `checksums.txt` file for SHA256 verification.

### Build from source

**Requirements:** Go 1.21 or later.

```bash
git clone https://github.com/matias2018/aura-devshield
cd aura-devshield
make install          # builds and copies to /usr/local/bin
# or
make build            # builds ./aura-devshield in the current directory
```

---

## Usage

### `scan` — report findings

Scan all installed VS Code extensions and report findings:

```bash
aura-devshield scan
aura-devshield          # same — scan is the default
```

Machine-readable JSON output (for scripting or CI):

```bash
aura-devshield scan --json
```

On first run, every installed extension version is stamped with a `first_seen` timestamp. Re-run `scan` regularly to keep the quarantine clock moving.

---

### `apply` — enforce the quarantine policy

Pin extensions whose updates are within the quarantine window, and release pins for extensions that have cleared it. **Dry-run by default.**

```bash
aura-devshield apply             # preview what would change
aura-devshield apply --confirm   # write changes to VS Code settings.json
```

What `--confirm` does:

- Adds `"publisher.name": false` entries to `extensions.autoUpdate` in VS Code `settings.json` for quarantined extensions — disabling auto-update for those extensions only.
- Removes those entries once extensions clear the quarantine window.
- Your other VS Code settings are never touched.

---

### `clean` — remove junk

Remove old duplicate versions (keeping the highest semver), malformed extensions, and orphaned directories. **Dry-run by default.**

```bash
aura-devshield clean             # preview what would be removed
aura-devshield clean --confirm   # actually delete
```

---

## How the quarantine works

The first time you run `scan`, each installed extension version is recorded with a `first_seen` timestamp in `~/.aura-devshield/state.json`. Any version first seen less than `quarantine_days` ago (default: 7) is flagged as `vscode.update_in_quarantine`.

Running `apply --confirm` writes a per-extension auto-update block into VS Code's `settings.json`. When the quarantine window expires, the next `apply --confirm` removes the block and re-enables auto-update.

**Why this works:** The community typically detects and reports compromised extension versions within hours to days of publication. A 7-day quarantine means you benefit from community detection before the update reaches your machine. Three supply-chain attacks on VS Code extensions were observed in May 2026 alone.

---

## Configuration

Create `~/.aura-devshield/config.json` to override defaults:

```json
{
  "quarantine_days": 14
}
```

| Key | Default | Description |
|---|---|---|
| `quarantine_days` | `7` | Days a newly-seen version must age before being trusted |

If the file does not exist, defaults are used. No setup required.

---

## Finding types

| ID | Severity | Description |
|---|---|---|
| `vscode.update_in_quarantine` | Medium | Extension version first seen within the quarantine window — auto-update not yet trusted |
| `vscode.multiple_installed_versions` | Medium | Two or more versions of the same extension installed simultaneously |
| `vscode.invalid_metadata` | Low | Extension has an empty or unresolved placeholder display name |
| `vscode.orphaned_extension_directory` | Low | Directory in extensions folder contains no `package.json` |
| `vscode.symlinked_extension_directory` | Low | Extension directory is a symbolic link |

---

## State and config file locations

| Path | Purpose |
|---|---|
| `~/.aura-devshield/state.json` | First-seen timestamps and quarantine pin tracking |
| `~/.aura-devshield/config.json` | User configuration |

Both files are created automatically on first run. They are plain JSON and safe to inspect, back up, or edit manually.

To reset all first-seen timestamps (e.g. when intentionally re-evaluating all installed extensions):

```bash
rm ~/.aura-devshield/state.json
```

---

## VS Code settings path

The `apply` subcommand reads and writes the following file:

| Platform | Path |
|---|---|
| macOS | `~/Library/Application Support/Code/User/settings.json` |
| Linux | `~/.config/Code/User/settings.json` |
| Windows | `%APPDATA%\Code\User\settings.json` |

Back this file up before running `apply --confirm` for the first time.

---

## Design philosophy

- **Local-first** — operates entirely on the local filesystem. No outbound connections.
- **Privacy-focused** — no telemetry, no analytics, no account required.
- **Visibility before enforcement** — reports findings; you decide what to act on.
- **Dry-run by default** — `apply` and `clean` preview changes before doing anything destructive.
- **Zero dependencies** — pure Go standard library. A single static binary with no runtime requirements.

**Out of scope:** antivirus, kernel monitoring, behavioral analysis, EDR, silent auto-remediation, corporate SOC integrations.

---

## Planned scanner coverage

VS Code extensions are the first target. The architecture is designed to support:

- npm / `package-lock.json`
- GitHub Actions workflow pinning
- Composer / `composer.lock`
- pip / `requirements.txt`, `pyproject.toml`

See [devReadme.md](devReadme.md) for how to add a new scanner.
