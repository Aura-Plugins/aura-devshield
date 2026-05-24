package scanner

/*
This file defines the Finding struct, which represents a security finding or issue detected during the scanning process.
This gives us a reusable finding format for future scanners:
	VS Code scanner
	npm scanner
	Composer scanner
	GitHub Actions scanner
*/

type Severity string

const (
	SeverityLow    Severity = "Low"
	SeverityMedium Severity = "Medium"
	SeverityHigh   Severity = "High"
)

type Finding struct {
	ID          string            `json:"id"`
	Fingerprint string            `json:"fingerprint"`
	Severity    Severity          `json:"severity"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Target      string            `json:"target"`
	Path        string            `json:"path,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
