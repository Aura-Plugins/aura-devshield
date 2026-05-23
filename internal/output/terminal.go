package output

import (
	"fmt"

	"github.com/matias2018/aura-devshield/internal/scanner"
	"github.com/matias2018/aura-devshield/internal/vscode"
)

func PrintExtensions(extensions []*vscode.Extension) {
	fmt.Printf("\nParsed %d extensions:\n\n", len(extensions))

	for _, extension := range extensions {
		fmt.Printf(
			"%s | %s\n  Version: %s\n  Path: %s\n",
			extension.CanonicalID(),
			extension.DisplayName,
			extension.Version,
			extension.Path,
		)
	}
}

func PrintFindings(findings []scanner.Finding) {
	fmt.Printf("\nFindings: %d\n\n", len(findings))

	for _, finding := range findings {
		fmt.Printf("[%s] %s\n", finding.Severity, finding.Title)
		fmt.Printf("Target: %s\n", finding.Target)
		fmt.Printf("%s\n\n", finding.Description)
	}
}
