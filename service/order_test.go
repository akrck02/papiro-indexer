package service

import "testing"

func TestSplitOrder(t *testing.T) {
	cases := map[string]struct {
		order int
		name  string
	}{
		"01-get-started.md": {1, "get-started.md"},
		"02_installation":   {2, "installation"},
		"10 Notes.md":       {10, "Notes.md"},
		"2026-10-17.md":     {0, "2026-10-17.md"},
		"1-overview.md":     {1, "overview.md"},
		"123-days.md":       {0, "123-days.md"},
		"get-started.md":    {0, "get-started.md"},
	}

	for name, expected := range cases {
		order, clean := SplitOrder(name)
		if order != expected.order || clean != expected.name {
			t.Errorf("%s: got %d %s, expected %d %s", name, order, clean, expected.order, expected.name)
		}
	}
}

func TestCleanPath(t *testing.T) {
	if CleanPath("02-installation/01-local-testing", true) != "installation/local-testing" {
		t.Error("pages lose the prefixes of their folders and their own")
	}
	if CleanPath("02-installation/01-diagram.png", false) != "installation/01-diagram.png" {
		t.Error("files keep their own name, only their folders change")
	}
}

func TestLinksBetweenOrderedPages(t *testing.T) {
	if result := ResolveWikiUrl("01-local-testing.md", "02-installation", true); result != "./installation/local-testing" {
		t.Errorf("got %s", result)
	}
	if result := ResolveWikiUrl("../images/01-diagram.png", "02-installation", false); result != "./images/01-diagram.png" {
		t.Errorf("got %s", result)
	}
}
