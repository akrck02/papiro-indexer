package model

type IndexItem struct {
	Type  IndexItemType        `json:"type"`
	Path  string               `json:"path"`
	Files map[string]IndexItem `json:"files"`
}

type IndexItemType int

const (
	Directory IndexItemType = iota + 1
	File
)
