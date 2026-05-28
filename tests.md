# Aura DevShield — Testing Guide

## Running tests

```bash
go test ./...                  # all packages
go test ./internal/vscode/...  # vscode package only
go test -v ./internal/vscode/  # verbose: see each test name
go test -run TestCompareVersions ./internal/vscode/  # single test
go test -cover ./...           # with coverage summary
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out  # HTML report
```

Run `go vet ./...` before committing. It catches issues tests miss.

---

## Test file layout

Each package gets its own `_test.go` file alongside the code it tests. Use the same package name (white-box testing) so unexported functions like `compareVersions` and `pinnedBySettings` are accessible without exporting them.

```
internal/vscode/
    clean_test.go              ← compareVersions, FindCleanTargets
    quarantine_test.go         ← UpdateState, FindQuarantineFindings, QuarantinedExtensionIDs, formatDuration
    settings_test.go           ← pinnedBySettings, computeQuarantineResults, PreviewQuarantine, ApplyQuarantine
    metadata_findings_test.go  ← FindInvalidMetadataFindings
    findings_test.go           ← FindMultiVersionFindings
internal/state/
    state_test.go              ← Load, Save, RecordVSCodeExtension, FirstSeen, Pin/Unpin
internal/config/
    config_test.go             ← Load, defaults, invalid JSON
internal/scanner/
    fingerprint_test.go        ← GenerateFingerprint determinism and collision resistance
```

---

## Priority 1 — `compareVersions` (internal/vscode/clean.go)

This is pure logic with no I/O. Highest confidence, lowest effort. Test it exhaustively.

```go
// internal/vscode/clean_test.go
package vscode

import "testing"

func TestCompareVersions(t *testing.T) {
    cases := []struct {
        a, b string
        want int // positive: a>b, zero: equal, negative: a<b
    }{
        {"1.2.3", "1.2.2", 1},
        {"1.2.2", "1.2.3", -1},
        {"1.2.3", "1.2.3", 0},
        {"2.0.0", "1.9.9", 1},
        {"1.10.0", "1.9.0", 1},   // numeric comparison, not lexicographic
        {"1.0.0", "1.0", 0},      // different segment counts
        {"0.0.1", "0.0.0", 1},
        {"1.0.0", "2.0.0", -1},
    }

    for _, tc := range cases {
        got := compareVersions(tc.a, tc.b)
        if sign(got) != sign(tc.want) {
            t.Errorf("compareVersions(%q, %q) = %d, want sign %d", tc.a, tc.b, got, tc.want)
        }
    }
}

func sign(n int) int {
    if n > 0 { return 1 }
    if n < 0 { return -1 }
    return 0
}
```

**Edge cases to cover:** leading zeros (`01.2`), unequal segment counts (`1.0` vs `1.0.0`), the `1.10` vs `1.9` numeric-not-lexicographic case (this is the easy one to get wrong).

---

## Priority 2 — `pinnedBySettings` and `computeQuarantineResults` (internal/vscode/settings.go)

These are pure functions. They contain the correctness-critical logic for the `apply` command. No filesystem needed.

```go
// internal/vscode/settings_test.go
package vscode

import (
    "testing"
    "reflect"
)

func TestPinnedBySettings(t *testing.T) {
    t.Run("map form returns pinned IDs", func(t *testing.T) {
        settings := map[string]interface{}{
            "extensions.autoUpdate": map[string]interface{}{
                "publisher.ext": false,
                "another.ext":  true,  // true = not pinned, should be ignored
            },
        }
        got := pinnedBySettings(settings)
        if !got["publisher.ext"] { t.Error("expected publisher.ext to be pinned") }
        if got["another.ext"] { t.Error("another.ext should not be pinned") }
    })

    t.Run("global bool form leaves nothing pinned", func(t *testing.T) {
        settings := map[string]interface{}{
            "extensions.autoUpdate": false,
        }
        got := pinnedBySettings(settings)
        if len(got) != 0 { t.Errorf("expected no pins, got %v", got) }
    })

    t.Run("missing key returns empty", func(t *testing.T) {
        got := pinnedBySettings(map[string]interface{}{})
        if len(got) != 0 { t.Errorf("expected no pins, got %v", got) }
    })
}

func TestComputeQuarantineResults(t *testing.T) {
    t.Run("pins new extension", func(t *testing.T) {
        settings := map[string]interface{}{}
        results, newPinned := computeQuarantineResults(settings, []string{"pub.ext"}, nil)
        if len(results) != 1 || results[0].Action != "pinned" {
            t.Errorf("unexpected results: %v", results)
        }
        if !newPinned["pub.ext"] {
            t.Error("pub.ext should be in newPinned")
        }
    })

    t.Run("releases cleared extension", func(t *testing.T) {
        settings := map[string]interface{}{
            "extensions.autoUpdate": map[string]interface{}{
                "pub.ext": false,
            },
        }
        results, newPinned := computeQuarantineResults(settings, nil, []string{"pub.ext"})
        if len(results) != 1 || results[0].Action != "released" {
            t.Errorf("unexpected results: %v", results)
        }
        if newPinned["pub.ext"] {
            t.Error("pub.ext should be removed from newPinned")
        }
    })

    t.Run("preserves manually pinned extensions not in toRelease", func(t *testing.T) {
        settings := map[string]interface{}{
            "extensions.autoUpdate": map[string]interface{}{
                "manual.pin": false,
                "manual.enabled": true, // explicitly enabled — must survive too
            },
        }
        _, newAutoUpdate := computeQuarantineResults(settings, []string{"quarantined.ext"}, nil)
        if newAutoUpdate["manual.pin"] != false {
            t.Error("manually pinned extension should be preserved")
        }
        if newAutoUpdate["manual.enabled"] != true {
            t.Error("manually enabled extension entry should be preserved")
        }
        if newAutoUpdate["quarantined.ext"] != false {
            t.Error("newly quarantined extension should be pinned")
        }
    })
}
```

**The `manual.pin` preservation test will fail against v0.3.0** — that is intentional. It documents the known bug described in the analysis. Fix `computeQuarantineResults` to start from all currently-pinned extensions (not just an empty map) and the test will pass.

---

## Priority 3 — `FindQuarantineFindings` (internal/vscode/quarantine.go)

Needs a fake state and controlled clock. Use `time.Now().Add(-duration)` to simulate ages.

```go
// internal/vscode/quarantine_test.go
package vscode

import (
    "testing"
    "time"

    "github.com/Aura-Plugins/aura-devshield/internal/state"
)

func TestFindQuarantineFindings(t *testing.T) {
    ext := &Extension{Publisher: "pub", Name: "ext", Version: "1.0.0"}
    delay := 7 * 24 * time.Hour

    t.Run("extension within window is flagged", func(t *testing.T) {
        st := &state.State{VSCodeExtensions: make(map[string]map[string]time.Time)}
        st.RecordVSCodeExtension(ext.CanonicalID(), ext.Version, time.Now().Add(-48*time.Hour))

        findings := FindQuarantineFindings([]*Extension{ext}, st, delay)
        if len(findings) != 1 {
            t.Fatalf("expected 1 finding, got %d", len(findings))
        }
        if findings[0].ID != "vscode.update_in_quarantine" {
            t.Errorf("unexpected finding ID: %s", findings[0].ID)
        }
    })

    t.Run("extension past window is not flagged", func(t *testing.T) {
        st := &state.State{VSCodeExtensions: make(map[string]map[string]time.Time)}
        st.RecordVSCodeExtension(ext.CanonicalID(), ext.Version, time.Now().Add(-8*24*time.Hour))

        findings := FindQuarantineFindings([]*Extension{ext}, st, delay)
        if len(findings) != 0 {
            t.Errorf("expected no findings, got %d", len(findings))
        }
    })

    t.Run("unknown version produces no finding", func(t *testing.T) {
        st := &state.State{VSCodeExtensions: make(map[string]map[string]time.Time)}
        findings := FindQuarantineFindings([]*Extension{ext}, st, delay)
        if len(findings) != 0 {
            t.Errorf("expected no findings for unknown version, got %d", len(findings))
        }
    })
}

func TestFormatDuration(t *testing.T) {
    cases := []struct {
        d    time.Duration
        want string
    }{
        {7 * 24 * time.Hour, "7 days"},
        {1 * 24 * time.Hour, "1 day"},
        {48 * time.Hour, "2 days"},
        {72 * time.Hour, "3 days"},
        {36 * time.Hour, "36h"},     // not a whole day
        {0, "0s"},
    }
    for _, tc := range cases {
        got := formatDuration(tc.d)
        if got != tc.want {
            t.Errorf("formatDuration(%v) = %q, want %q", tc.d, got, tc.want)
        }
    }
}
```

---

## Priority 4 — `state.Load` and `state.Save` (internal/state/state.go)

Use `t.TempDir()` for filesystem tests — Go cleans it up automatically after the test.

```go
// internal/state/state_test.go
package state

import (
    "path/filepath"
    "testing"
    "time"
)

func TestLoadMissingFile(t *testing.T) {
    s, err := Load(filepath.Join(t.TempDir(), "state.json"))
    if err != nil { t.Fatal(err) }
    if s.VSCodeExtensions == nil { t.Error("VSCodeExtensions should be initialised") }
    if s.VSCodePinned == nil { t.Error("VSCodePinned should be initialised") }
}

func TestRoundTrip(t *testing.T) {
    path := filepath.Join(t.TempDir(), "state.json")
    s := &State{
        VSCodeExtensions: make(map[string]map[string]time.Time),
        VSCodePinned:     make(map[string]bool),
    }

    now := time.Now().UTC().Truncate(time.Second)
    s.RecordVSCodeExtension("pub.ext", "1.0.0", now)
    s.PinVSCodeExtension("pub.ext")

    if err := s.Save(path); err != nil { t.Fatal(err) }

    loaded, err := Load(path)
    if err != nil { t.Fatal(err) }

    got, ok := loaded.FirstSeen("pub.ext", "1.0.0")
    if !ok { t.Fatal("first_seen not found after reload") }
    if !got.Equal(now) { t.Errorf("first_seen: got %v, want %v", got, now) }

    if !loaded.IsVSCodeExtensionPinned("pub.ext") {
        t.Error("extension should still be pinned after reload")
    }
}

func TestRecordIsImmutable(t *testing.T) {
    s := &State{VSCodeExtensions: make(map[string]map[string]time.Time)}
    t1 := time.Now().Add(-24 * time.Hour).UTC()
    t2 := time.Now().UTC()

    s.RecordVSCodeExtension("pub.ext", "1.0.0", t1)
    s.RecordVSCodeExtension("pub.ext", "1.0.0", t2) // second call must be a no-op

    got, _ := s.FirstSeen("pub.ext", "1.0.0")
    if !got.Equal(t1) {
        t.Errorf("first_seen was overwritten: got %v, want %v", got, t1)
    }
}
```

---

## Priority 5 — `ApplyQuarantine` with temp files (internal/vscode/settings.go)

```go
// internal/vscode/settings_test.go (append to existing file)

func TestApplyQuarantine(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "settings.json")

    t.Run("pins extension in empty settings", func(t *testing.T) {
        results, err := ApplyQuarantine(path, []string{"pub.ext"}, nil)
        if err != nil { t.Fatal(err) }
        if len(results) != 1 || results[0].Action != "pinned" {
            t.Errorf("unexpected results: %v", results)
        }

        settings, _ := readSettings(path)
        m, ok := settings["extensions.autoUpdate"].(map[string]interface{})
        if !ok { t.Fatal("extensions.autoUpdate should be a map") }
        if m["pub.ext"] != false {
            t.Errorf("pub.ext should be false, got %v", m["pub.ext"])
        }
    })

    t.Run("releases extension removes entry", func(t *testing.T) {
        results, err := ApplyQuarantine(path, nil, []string{"pub.ext"})
        if err != nil { t.Fatal(err) }
        if len(results) != 1 || results[0].Action != "released" {
            t.Errorf("unexpected results: %v", results)
        }

        settings, _ := readSettings(path)
        if _, exists := settings["extensions.autoUpdate"]; exists {
            t.Error("extensions.autoUpdate should be removed when map is empty")
        }
    })
}

// CLI integration note: apply and clean now run for real by default.
// Pass --dry-run to preview without writing. Tests above target the library
// layer directly and are unaffected by the flag change.
```

Note: `readSettings` is unexported, so this test must be in `package vscode` (same package, not `package vscode_test`).

---

## Priority 6 — `config.Load` (internal/config/config.go)

```go
// internal/config/config_test.go
package config

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoadMissingUsesDefaults(t *testing.T) {
    c, err := Load(filepath.Join(t.TempDir(), "config.json"))
    if err != nil { t.Fatal(err) }
    if c.QuarantineDays != DefaultQuarantineDays {
        t.Errorf("QuarantineDays: got %d, want %d", c.QuarantineDays, DefaultQuarantineDays)
    }
}

func TestLoadZeroQuarantineDefaultsToSeven(t *testing.T) {
    path := filepath.Join(t.TempDir(), "config.json")
    os.WriteFile(path, []byte(`{"quarantine_days": 0}`), 0644)

    c, err := Load(path)
    if err != nil { t.Fatal(err) }
    if c.QuarantineDays != DefaultQuarantineDays {
        t.Errorf("expected default, got %d", c.QuarantineDays)
    }
}

func TestLoadInvalidJSONErrors(t *testing.T) {
    path := filepath.Join(t.TempDir(), "config.json")
    os.WriteFile(path, []byte(`not json`), 0644)

    _, err := Load(path)
    if err == nil { t.Error("expected error for invalid JSON") }
}
```

---

## Recommended test helpers

Add a small `testutil` helper in each package's test file to reduce boilerplate when constructing `Extension` values:

```go
func ext(publisher, name, version string) *Extension {
    return &Extension{Publisher: publisher, Name: name, Version: version}
}
```

And a fresh `State`:

```go
func emptyState() *state.State {
    return &state.State{
        VSCodeExtensions: make(map[string]map[string]time.Time),
        VSCodePinned:     make(map[string]bool),
    }
}
```

---

## What not to test

- `output/` (terminal and JSON formatters) — output formatting is better verified by running the CLI directly. Unit tests here would just assert on exact string output and break on every label change.
- `vscode.ScanExtensions` and `vscode.ReadPackageJSON` — these are thin wrappers around `os.ReadDir` and `json.Unmarshal`. Testing them means creating fixture directories; not worth it until you have a broader integration test suite.
- `vscode.ExtensionsDir` and `vscode.VSCodeSettingsPath` — pure path construction based on `os.UserHomeDir()`. Asserting a specific path string would tie tests to the CI runner's home directory.

---

## CI integration

Once tests exist, add a test step to `.github/workflows/release.yml` before the build:

```yaml
- name: Test
  run: go test ./...
```

Place it after the `Vet` step and before `Build all targets`. The workflow already runs on tag pushes — tests there act as a final gate before a release binary is published.

For local pre-commit enforcement, add to `Makefile`:

```makefile
## test: run all tests
test:
	go test ./...

## check: vet + test (run before committing)
check: vet test
```
