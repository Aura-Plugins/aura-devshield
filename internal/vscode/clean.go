package vscode

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// CleanTarget is a directory that should be removed and why.
type CleanTarget struct {
	Path     string
	Reason   string
	KeepPath string // set for duplicate versions; empty otherwise
}

// FindCleanTargets returns directories that are safe to remove:
// older duplicate versions (keeping the highest semver), malformed extensions
// (missing name or publisher), and orphaned directories (no package.json).
func FindCleanTargets(extensions []*Extension, extensionsDir string) ([]CleanTarget, error) {
	seen := make(map[string]bool)
	var targets []CleanTarget

	addTarget := func(t CleanTarget) {
		if !seen[t.Path] {
			seen[t.Path] = true
			targets = append(targets, t)
		}
	}

	// Duplicate versions: group by canonical ID, keep highest semver.
	byID := make(map[string][]*Extension)
	for _, ext := range extensions {
		byID[ext.CanonicalID()] = append(byID[ext.CanonicalID()], ext)
	}

	for _, group := range byID {
		if len(group) <= 1 {
			continue
		}
		sort.Slice(group, func(i, j int) bool {
			return compareVersions(group[i].Version, group[j].Version) > 0
		})
		keep := group[0]
		for _, ext := range group[1:] {
			addTarget(CleanTarget{
				Path:     ext.Path,
				Reason:   fmt.Sprintf("duplicate version %s (keeping %s)", ext.Version, keep.Version),
				KeepPath: keep.Path,
			})
		}
	}

	// Malformed: parsed successfully but missing name or publisher.
	for _, ext := range extensions {
		if ext.Name == "" || ext.Publisher == "" {
			addTarget(CleanTarget{
				Path:   ext.Path,
				Reason: "malformed: missing name or publisher in package.json",
			})
		}
	}

	// Orphaned: directories with no package.json at all.
	orphaned, err := FindOrphanedExtensionDirs(extensionsDir)
	if err != nil {
		return nil, err
	}
	for _, dir := range orphaned {
		addTarget(CleanTarget{
			Path:   dir,
			Reason: "orphaned: no package.json found",
		})
	}

	return targets, nil
}

// Clean removes each target directory. Call only after user confirmation.
func Clean(targets []CleanTarget) error {
	for _, t := range targets {
		if err := os.RemoveAll(t.Path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", t.Path, err)
		}
	}
	return nil
}

// compareVersions compares two dot-separated version strings.
// Returns positive if a > b, zero if equal, negative if a < b.
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var ap, bp string
		if i < len(aParts) {
			ap = aParts[i]
		}
		if i < len(bParts) {
			bp = bParts[i]
		}

		an, aerr := strconv.Atoi(ap)
		bn, berr := strconv.Atoi(bp)

		if aerr == nil && berr == nil {
			if an != bn {
				if an > bn {
					return 1
				}
				return -1
			}
		} else {
			if ap != bp {
				if ap > bp {
					return 1
				}
				return -1
			}
		}
	}
	return 0
}
