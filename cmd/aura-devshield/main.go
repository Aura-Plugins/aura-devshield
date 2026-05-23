package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/matias2018/aura-devshield/internal/output"
	"github.com/matias2018/aura-devshield/internal/scanner"
	"github.com/matias2018/aura-devshield/internal/vscode"
)

func main() {
	jsonOutput := flag.Bool("json", false, "Output findings as JSON")
	flag.Parse()

	if !*jsonOutput {
		fmt.Println("Starting... Aura DevShield!")
	}

	extensionsDir, err := vscode.ExtensionsDir()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if !*jsonOutput {
		fmt.Println("VS Code extensions directory:")
		fmt.Println(extensionsDir)
	}

	info, err := os.Stat(extensionsDir)
	if err != nil {
		fmt.Println("Directory does not exist or cannot be accessed")
		return
	}

	if !info.IsDir() {
		fmt.Println("Path exists but is not a directory")
		return
	}

	if !*jsonOutput {
		fmt.Println("Directory exists")
	}

	extensions, err := vscode.ScanExtensions(extensionsDir)
	if err != nil {
		fmt.Println("Error scanning extensions:", err)
		return
	}

	var findings []scanner.Finding

	findings = append(
		findings,
		vscode.FindMultiVersionFindings(extensions)...,
	)

	findings = append(
		findings,
		vscode.FindInvalidMetadataFindings(extensions)...,
	)

	// Keep instance-level findings for now.
	// Deduplication may be exposed later through a --dedupe flag.
	// findings = scanner.DeduplicateFindings(findings)

	if *jsonOutput {
		if err := output.PrintFindingsJSON(findings); err != nil {
			fmt.Println("Error writing JSON:", err)
			return
		}

		return
	}

	output.PrintExtensions(extensions)
	output.PrintFindings(findings)
}
