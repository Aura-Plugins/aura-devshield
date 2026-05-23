package vscode

func FindMultipleInstalledVersions(extensions []*Extension) map[string][]*Extension {
	byID := make(map[string][]*Extension)

	for _, extension := range extensions {
		id := extension.CanonicalID()
		byID[id] = append(byID[id], extension)
	}

	duplicates := make(map[string][]*Extension)

	for id, groupedExtensions := range byID {
		if len(groupedExtensions) > 1 {
			duplicates[id] = groupedExtensions
		}
	}

	return duplicates
}
