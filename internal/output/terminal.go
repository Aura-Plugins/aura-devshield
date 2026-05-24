package output

import (
	"fmt"

	"github.com/matias2018/aura-devshield/internal/scanner"
	"github.com/matias2018/aura-devshield/internal/vscode"
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

func PrintQuarantinePolicy(days int) {
	fmt.Printf("\n  Quarantine policy: %s\n",
		col(fmt.Sprintf("%d days", days), ansiBold),
	)
}

func PrintQuarantineResults(results []vscode.QuarantineResult, settingsPath string, applied bool) {
	if len(results) == 0 {
		fmt.Printf("\n  %s\n\n", col("No quarantine changes needed.", ansiGray))
		return
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

	if applied {
		fmt.Printf("  %s\n  %s\n\n",
			col("Applied to:", ansiGray),
			col("  "+settingsPath, ansiDim),
		)
	} else {
		fmt.Printf("  %s\n  %s\n\n",
			col("Dry run — run 'aura-devshield apply --confirm' to write to:", ansiGray),
			col("  "+settingsPath, ansiDim),
		)
	}
}

func PrintCleanTargets(targets []vscode.CleanTarget, applied bool) {
	if len(targets) == 0 {
		fmt.Printf("\n  %s\n\n", col("✓  Nothing to clean.", ansiBoldGreen))
		return
	}

	n := len(targets)
	noun := "directories"
	if n == 1 {
		noun = "directory"
	}

	if applied {
		fmt.Printf("\n  %s\n\n",
			col(fmt.Sprintf("Removed %d %s.", n, noun), ansiBoldRed),
		)
	} else {
		fmt.Printf("\n  %s\n\n",
			col(fmt.Sprintf("Would remove %d %s:", n, noun), ansiBold),
		)
	}

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

	if !applied {
		fmt.Printf("  %s\n\n",
			col("Run 'aura-devshield clean --confirm' to delete.", ansiGray),
		)
	}
}
