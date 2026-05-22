package vscode

type Extension struct {
	Name   string `json:"name"`
	DisplayName string `json:"displayName"`
	Version string `json:"version"`
	Publisher string `json:"publisher"`
	Description string `json:"description"`
	Path string
}

func (e Extension) ID() string {
	return e.Publisher + "." + e.Name
}