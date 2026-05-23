# Aura DevShield — Specs & Features

## Project Overview

Aura DevShield is a developer-focused local security visibility tool.

Its primary goal is to help developers inspect, understand and reduce risks in their local development environments.

---

# Core Design Principles

- Local-first
- Privacy-focused
- No mandatory cloud dependency
- No telemetry by default
- Human-readable output
- Machine-readable output
- Security visibility before enforcement
- Safe and transparent operation

---

# Aura DevShield 1.0.0 — Planned Features

# 1. VS Code Extension Discovery

## Features

- Scan installed VS Code extensions
- Parse package.json metadata
- Discover extension publishers
- Discover extension versions
- Discover extension descriptions
- Detect extension installation paths
- Normalize extension identifiers

---

# 2. Canonical Extension ID System

## Features

Normalize extension identifiers into a stable format.

Example:

- GitHub.copilot-chat
→ github.copilot-chat

---

# 3. Multi-Version Extension Detection

## Features

Detect extensions with multiple simultaneously-installed versions.

Example:

- anthropic.claude-code
  - 2.1.141
  - 2.1.142
  - 2.1.143
  - 2.1.144
  - 2.1.145

## Security Relevance

Potential detection of:

- rollback attack surface
- stale vulnerable versions
- dormant malicious persistence
- abandoned extension remnants
- forensic residue
- edge-case activation risks

---

# 4. Suspicious Extension Detection

## Metadata Anomalies

Potential checks:

- Missing publisher
- Empty display names
- Broken placeholders
- Invalid metadata
- Missing package.json

Examples:

- %displayName%
- %extension.title%

## Installation Anomalies

Potential checks:

- Unexpected installation paths
- Hidden folders
- Suspicious folder naming
- Duplicate installations

## Content Anomalies

Potential checks:

- Embedded executables
- Native binaries
- Shell scripts
- Suspicious permissions
- Obfuscated blobs

---

# 5. Structured Findings Engine

## Features

Generate structured findings containing:

- ID
- severity
- title
- description
- target

---

# 6. Severity Classification System

## Levels

- Low
- Medium
- High
- Critical

---

# 7. JSON Output Mode

## Features

CLI JSON export mode:

```bash
aura-devshield --json
```

---

# 8. Extension Cleanup Suggestions

## Features

Suggest cleanup of:

- old versions
- abandoned versions
- stale extension installations
- duplicate installs

---

# 9. Interactive Cleanup Mode

## Features

Allow users to:

- select extensions
- remove old versions
- keep latest version only
- perform guided cleanup

Example:

```bash
aura-devshield clean
```

---

# 10. Delayed Extension Update System

## Features

Allow users to delay automatic extension updates by:

- 12h
- 24h
- 48h
- custom durations

## Goal

Create a security cooling-off period for newly released extension versions.

---

# 11. Extension Installation Timeline

## Features

Track:

- installation dates
- update dates
- newest extensions
- oldest extensions

---

# 12. Publisher Inventory

## Features

Show:

- publishers with most extensions
- unknown publishers
- abandoned publishers
- AI tooling inventory

---

# 13. AI Tooling Detection

## Features

Specifically identify AI-related tooling:

- Claude Code
- GitHub Copilot
- Cody
- Tabnine
- MCP tooling
- AI-assisted extensions

---

# 14. Human-Friendly CLI Output

## Features

Readable terminal output.

Optional future support:

- colors
- grouped findings
- summaries

---

# 15. Machine-Friendly Exit Codes

## Features

```text
0 = clean
1 = findings detected
2 = critical findings detected
```

---

# 16. Local-Only Operation

## Features

By default:

- no telemetry
- no cloud dependency
- no account required
- no remote uploads

---

# Features Explicitly Out of Scope for 1.0.0

The following are intentionally NOT part of the initial release:

- Kernel monitoring
- EDR functionality
- Behavioral analysis
- Antivirus signatures
- Silent auto-remediation
- Corporate SOC integrations
- Process injection
- Real-time blocking engines

---

# Recommended 1.0.0 MVP Scope

## Core MVP

- VS Code extension scanning
- Metadata parsing
- Canonical IDs
- Multi-version detection
- Structured findings
- Severity system
- JSON export
- Cleanup suggestions
- Safe local-only CLI operation

## Optional Stretch Goal

- Delayed extension update / quarantine system

---

# Long-Term Vision

Aura DevShield may eventually evolve into:

- Developer workstation auditing platform
- Developer supply-chain visibility layer
- Local tooling inspection utility
- Security observability tool for developer environments

without abandoning its core philosophy:

> Visibility before enforcement.
