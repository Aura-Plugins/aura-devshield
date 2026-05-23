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
	SeverityLow	Severity = "Low"
	SeverityMedium Severity = "Medium"
	SeverityHigh Severity = "High"
)

type Finding struct {
	ID          string
	Severity    Severity
	Title       string
	Description string
	Target      string
}