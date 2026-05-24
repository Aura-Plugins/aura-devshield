package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Aura-Plugins/aura-devshield/internal/config"
	"github.com/Aura-Plugins/aura-devshield/internal/output"
	"github.com/Aura-Plugins/aura-devshield/internal/scanner"
	"github.com/Aura-Plugins/aura-devshield/internal/state"
	"github.com/Aura-Plugins/aura-devshield/internal/vscode"
)

// version is set at build time via -ldflags "-X main.version=v1.0.0".
var version = "dev"

func main() {
	args := os.Args[1:]

	if len(args) == 1 && (args[0] == "--version" || args[0] == "-version" || args[0] == "version") {
		fmt.Println("aura-devshield", version)
		return
	}

	// Backwards-compatible: no args or first arg is a flag → default to scan.
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		runScan(args)
		return
	}

	switch args[0] {
	case "scan":
		runScan(args[1:])
	case "apply":
		runApply(args[1:])
	case "clean":
		runClean(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\nUsage: aura-devshield [scan|apply|clean] [flags]\n\n  scan  [--json]      Report all findings (default)\n  apply [--confirm]   Pin/release extensions in VS Code settings\n  clean [--confirm]   Remove duplicate and malformed extensions\n", args[0])
		os.Exit(1)
	}
}

// runScan runs all scanners and prints findings.
func runScan(args []string) {
	flags := flag.NewFlagSet("scan", flag.ExitOnError)
	jsonOutput := flags.Bool("json", false, "Output findings as JSON")
	list := flags.Bool("list", false, "Also print all installed extensions")
	flags.Parse(args)

	ss, err := prepare()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	if err := ss.st.Save(ss.statePath); err != nil && !*jsonOutput {
		fmt.Fprintf(os.Stderr, "Warning: could not save state: %v\n", err)
	}

	var findings []scanner.Finding

	findings = append(findings, vscode.FindMultiVersionFindings(ss.extensions)...)
	findings = append(findings, vscode.FindInvalidMetadataFindings(ss.extensions)...)
	findings = append(findings, vscode.FindQuarantineFindings(ss.extensions, ss.st, ss.cfg.QuarantineDays)...)

	symlinkFindings, err := vscode.FindSymlinkedExtensionDirectoryFindings(ss.extensionsDir)
	if err != nil && !*jsonOutput {
		fmt.Fprintf(os.Stderr, "Warning: symlink scan: %v\n", err)
	} else {
		findings = append(findings, symlinkFindings...)
	}

	orphanFindings, err := vscode.FindOrphanedDirectoryFindings(ss.extensionsDir)
	if err != nil && !*jsonOutput {
		fmt.Fprintf(os.Stderr, "Warning: orphan scan: %v\n", err)
	} else {
		findings = append(findings, orphanFindings...)
	}

	if *jsonOutput {
		if err := output.PrintFindingsJSON(findings); err != nil {
			fmt.Fprintln(os.Stderr, "Error writing JSON:", err)
			os.Exit(1)
		}
		return
	}

	if *list {
		output.PrintExtensions(ss.extensions)
	}
	output.PrintFindings(findings, ss.extensionsDir, len(ss.extensions))
}

// runApply pins quarantined extensions and releases cleared ones in VS Code settings.json.
func runApply(args []string) {
	flags := flag.NewFlagSet("apply", flag.ExitOnError)
	confirm := flags.Bool("confirm", false, "Write changes to VS Code settings.json")
	flags.Parse(args)

	ss, err := prepare()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	settingsPath, err := vscode.VSCodeSettingsPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error resolving VS Code settings path:", err)
		os.Exit(1)
	}

	quarantined := vscode.QuarantinedExtensionIDs(ss.extensions, ss.st, ss.cfg.QuarantineDays)

	quarantinedSet := make(map[string]bool, len(quarantined))
	for _, id := range quarantined {
		quarantinedSet[id] = true
	}

	var toPin, toRelease []string
	for _, id := range quarantined {
		if !ss.st.IsVSCodeExtensionPinned(id) {
			toPin = append(toPin, id)
		}
	}
	for _, id := range ss.st.PinnedVSCodeExtensions() {
		if !quarantinedSet[id] {
			toRelease = append(toRelease, id)
		}
	}

	output.PrintQuarantinePolicy(ss.cfg.QuarantineDays)

	if *confirm {
		results, err := vscode.ApplyQuarantine(settingsPath, toPin, toRelease)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error applying quarantine:", err)
			os.Exit(1)
		}

		for _, id := range toPin {
			ss.st.PinVSCodeExtension(id)
		}
		for _, id := range toRelease {
			ss.st.UnpinVSCodeExtension(id)
		}
		if err := ss.st.Save(ss.statePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save state: %v\n", err)
		}

		output.PrintQuarantineResults(results, settingsPath, true)
	} else {
		results, err := vscode.PreviewQuarantine(settingsPath, toPin, toRelease)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error previewing quarantine:", err)
			os.Exit(1)
		}
		output.PrintQuarantineResults(results, settingsPath, false)
	}
}

// runClean removes duplicate versions, malformed extensions, and orphaned directories.
func runClean(args []string) {
	flags := flag.NewFlagSet("clean", flag.ExitOnError)
	confirm := flags.Bool("confirm", false, "Actually remove directories")
	flags.Parse(args)

	extensionsDir, err := vscode.ExtensionsDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	extensions, err := vscode.ScanExtensions(extensionsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error scanning extensions:", err)
		os.Exit(1)
	}

	targets, err := vscode.FindCleanTargets(extensions, extensionsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error finding clean targets:", err)
		os.Exit(1)
	}

	if *confirm {
		if err := vscode.Clean(targets); err != nil {
			fmt.Fprintln(os.Stderr, "Error cleaning:", err)
			os.Exit(1)
		}
	}

	output.PrintCleanTargets(targets, *confirm)
}

// scanState holds everything needed to run findings across subcommands.
type scanState struct {
	cfg          *config.Config
	st           *state.State
	statePath    string
	extensionsDir string
	extensions   []*vscode.Extension
}

// prepare loads config and state, scans extensions, and records first-seen timestamps.
func prepare() (*scanState, error) {
	cfgPath, err := config.DefaultPath()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}

	statePath, err := state.DefaultPath()
	if err != nil {
		return nil, err
	}
	st, err := state.Load(statePath)
	if err != nil {
		return nil, err
	}

	extensionsDir, err := vscode.ExtensionsDir()
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(extensionsDir)
	if err != nil {
		return nil, fmt.Errorf("extensions directory inaccessible: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", extensionsDir)
	}

	extensions, err := vscode.ScanExtensions(extensionsDir)
	if err != nil {
		return nil, err
	}

	vscode.UpdateState(extensions, st)

	return &scanState{
		cfg:          cfg,
		st:           st,
		statePath:    statePath,
		extensionsDir: extensionsDir,
		extensions:   extensions,
	}, nil
}
