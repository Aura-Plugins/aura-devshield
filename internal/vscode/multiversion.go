package vscode

func FindMultiVersionExtensions(extensions []*Extension) map[string][]*Extension {
	byID := make(map[string][]*Extension)

	for _, extension := range extensions {
		id := extension.CanonicalID()
		byID[id] = append(byID[id], extension)
	}

	multiVersionExtensions := make(map[string][]*Extension)

	for id, groupedExtensions := range byID {
		if len(groupedExtensions) > 1 {
			multiVersionExtensions[id] = groupedExtensions
		}
	}

	return multiVersionExtensions
}