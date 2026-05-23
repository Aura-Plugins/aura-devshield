package main

import (
	"fmt"
	"os"

	"github.com/matias2018/aura-devshield/internal/vscode"
	"github.com/matias2018/aura-devshield/internal/output"
)

func main() {
	fmt.Println("Hello, Aura DevShield!")

	extensionsDir, err := vscode.ExtensionsDir()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("VS Code extensions directory:")
	fmt.Println(extensionsDir)

	info, err := os.Stat(extensionsDir)
	if err != nil {
		fmt.Println("Directory does not exist or cannot be accessed")
		return
	}

	if !info.IsDir() {
		fmt.Println("Path exists but is not a directory")
		return
	}

	fmt.Println("Directory exists")

	extensions, err := vscode.ScanExtensions(extensionsDir)
	if err != nil {
		fmt.Println("Error scanning extensions:", err)
		return
	}

	output.PrintExtensions(extensions)
	for _, extension := range extensions {
		fmt.Printf(
			"%s | %s\n  Version: %s\n  Path: %s\n",
			extension.CanonicalID(),
			extension.DisplayName,
			extension.Version,
			extension.Path,
		)
	}

	findings := vscode.FindMultiVersionFindings(extensions)

	output.PrintFindings(findings)
	for _, finding := range findings {
		fmt.Printf("[%s] %s\n", finding.Severity, finding.Title)
		fmt.Printf("Target: %s\n", finding.Target)
		fmt.Printf("%s\n\n", finding.Description)
	}
}