package model

type IndexItem struct {
	Type  IndexItemType        `json:"type"`
	Path  string               `json:"path"`
	Files map[string]IndexItem `json:"files"`
	// position in its folder, from a "01-" prefix in its name; 0 when it has none
	Order int `json:"order,omitempty"`
}

type IndexItemType int

const (
	Directory IndexItemType = iota + 1
	File
)
