package service

import (
	"strings"
	"testing"
)

func TestAddPageSearch(t *testing.T) {
	defer func() { searchPages = map[string]SearchPage{} }()
	AddPageSearch("guides/layout", []byte("# Layout guide\n\nBoxes **grow** and `Fit`.\n\n## Sizes\n\nText.\n\n```go\nui.Row()\n```\n"))

	page := searchPages["guides/layout"]
	if "Layout guide" != page.Title {
		t.Errorf("title: %q", page.Title)
	}
	if 1 != len(page.Sections) || "sizes" != page.Sections[0].ID || "Sizes" != page.Sections[0].Text {
		t.Errorf("sections: %+v", page.Sections)
	}
	for _, expected := range []string{"Boxes grow and Fit", "ui.Row()"} {
		if !strings.Contains(page.Text, expected) {
			t.Errorf("%q not in %q", expected, page.Text)
		}
	}
}
