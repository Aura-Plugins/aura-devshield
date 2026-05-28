package output

import (
	"fmt"
	"time"

	"github.com/Aura-Plugins/aura-devshield/internal/scanner"
	"github.com/Aura-Plugins/aura-devshield/internal/vscode"
)

// PrintFindings renders all findings in the TUI style, grouped High → Medium → Low.
// It also prints the scan header (directory + extension count).
func PrintFindings(findings []scanner.Finding, extensionsDir string, extensionCount int) {
	tuiHeader(extensionsDir, extensionCount)

	var high, medium, low int
	for _, f := range findings {
		switch f.Severity {
		case scanner.SeverityHigh:
			high++
		case scanner.SeverityMedium:
			medium++
		case scanner.SeverityLow:
			low++
		}
	}

	tuiSummary(len(findings), high, medium, low)
	if len(findings) == 0 {
		return
	}

	for _, sev := range []scanner.Severity{
		scanner.SeverityHigh,
		scanner.SeverityMedium,
		scanner.SeverityLow,
	} {
		var group []scanner.Finding
		for _, f := range findings {
			if f.Severity == sev {
				group = append(group, f)
			}
		}
		if len(group) == 0 {
			continue
		}

		var label string
		switch sev {
		case scanner.SeverityHigh:
			label = "HIGH"
		case scanner.SeverityMedium:
			label = "MEDIUM"
		default:
			label = "LOW"
		}
		tuiSectionHeader(label, sev)
		for _, f := range group {
			tuiFinding(f)
		}
	}

	tuiDivider("")
	fmt.Println()
}

// PrintExtensions renders the full extension inventory (used with --list flag).
func PrintExtensions(extensions []*vscode.Extension) {
	fmt.Printf("\n  %s\n\n", col(fmt.Sprintf("%d extensions installed", len(extensions)), ansiBold))
	for _, ext := range extensions {
		fmt.Printf("  %s  %s  %s\n",
			col(ext.CanonicalID(), ansiBold),
			col("@"+ext.Version, ansiGray),
			col(ext.DisplayName, ansiDim),
		)
	}
	fmt.Println()
}

func PrintQuarantinePolicy(delay time.Duration) {
	h := int(delay.Hours())
	var label string
	if h%24 == 0 {
		days := h / 24
		if days == 1 {
			label = "1 day"
		} else {
			label = fmt.Sprintf("%d days", days)
		}
	} else {
		label = fmt.Sprintf("%dh", h)
	}
	fmt.Printf("\n  Quarantine policy: %s\n", col(label, ansiBold))
}

// PrintQuarantineResults is used by the --dry-run path: shows the change list
// with a footer that tells the user what to run next.
func PrintQuarantineResults(results []vscode.QuarantineResult, settingsPath string, applied bool) {
	if !PrintQuarantinePlan(results) {
		return
	}
	if applied {
		PrintQuarantineApplied(settingsPath)
	} else {
		fmt.Printf("  %s\n  %s\n\n",
			col("Dry run — run 'aura-devshield apply' to write to:", ansiGray),
			col("  "+settingsPath, ansiDim),
		)
	}
}

// PrintQuarantinePlan prints the pin/release change list and divider with no
// footer, leaving room for a prompt or applied message. Returns true if there
// are changes to show.
func PrintQuarantinePlan(results []vscode.QuarantineResult) bool {
	if len(results) == 0 {
		fmt.Printf("\n  %s\n\n", col("No quarantine changes needed.", ansiGray))
		return false
	}

	fmt.Println()
	for _, r := range results {
		switch r.Action {
		case "pinned":
			fmt.Printf("  %s  %s\n",
				col("▲", ansiBoldYellow),
				col("Pin (disable auto-update):     "+r.ExtensionID, ansiBold),
			)
		case "released":
			fmt.Printf("  %s  %s\n",
				col("✓", ansiBoldGreen),
				col("Release (re-enable auto-update): "+r.ExtensionID, ansiBold),
			)
		}
	}

	fmt.Println()
	tuiDivider("")
	fmt.Println()
	return true
}

// PrintQuarantineApplied prints the "Applied to:" footer after a successful apply.
func PrintQuarantineApplied(settingsPath string) {
	fmt.Printf("  %s\n  %s\n\n",
		col("Applied to:", ansiGray),
		col("  "+settingsPath, ansiDim),
	)
}

// PrintCleanTargets is used by the --dry-run path: shows the target list with a
// footer that tells the user what to run next.
func PrintCleanTargets(targets []vscode.CleanTarget, applied bool) {
	if !PrintCleanPlan(targets) {
		return
	}
	if applied {
		// PrintCleanPlan showed "Would remove N dirs:" header; restate as applied.
		// (This path is not used after the interactive refactor but kept for compat.)
	} else {
		fmt.Printf("  %s\n\n",
			col("Run 'aura-devshield clean' to delete.", ansiGray),
		)
	}
}

// PrintCleanPlan prints "Would remove N directories:" followed by the target list
// and a divider, with no footer. Returns true if there are targets to show.
func PrintCleanPlan(targets []vscode.CleanTarget) bool {
	if len(targets) == 0 {
		fmt.Printf("\n  %s\n\n", col("✓  Nothing to clean.", ansiBoldGreen))
		return false
	}

	n := len(targets)
	noun := "directories"
	if n == 1 {
		noun = "directory"
	}

	fmt.Printf("\n  %s\n\n",
		col(fmt.Sprintf("Would remove %d %s:", n, noun), ansiBold),
	)

	for _, t := range targets {
		fmt.Printf("  %s  %s\n",
			col("×", ansiRed),
			col(t.Path, ansiBold),
		)
		fmt.Printf("     %s\n", col(t.Reason, ansiGray))
		if t.KeepPath != "" {
			fmt.Printf("     %s %s\n", col("keeping:", ansiGray), col(t.KeepPath, ansiDim))
		}
		fmt.Println()
	}

	tuiDivider("")
	fmt.Println()
	return true
}

// PrintCleanApplied prints the "Removed N directories." confirmation.
func PrintCleanApplied(n int) {
	noun := "directories"
	if n == 1 {
		noun = "directory"
	}
	fmt.Printf("\n  %s\n\n", col(fmt.Sprintf("Removed %d %s.", n, noun), ansiBoldGreen))
}

// PrintAborted prints the cancellation message shown when a user declines a prompt.
func PrintAborted() {
	fmt.Printf("\n  %s\n\n", col("Aborted.", ansiGray))
}
