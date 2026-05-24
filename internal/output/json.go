package output

import (
	"encoding/json"
	"github.com/Aura-Plugins/aura-devshield/internal/scanner"
	"os"
)

func PrintFindingsJSON(findings []scanner.Finding) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	return encoder.Encode(findings)
}
