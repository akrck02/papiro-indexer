package service

import "testing"

func TestMatchesQuery(t *testing.T) {
	page := newColorPage("Projects/Garden/Tomatoes.md", "---\ntype: plant\ntags: [summer, food/vegetable]\n---\nPlanted in #spring by the wall.\n")

	cases := map[string]bool{
		"path:Projects/Garden":          true,
		"path:Kitchen":                  false,
		"file:tomato":                   true,
		"tag:#summer":                   true,
		"tag:food":                      true,
		"tag:spring":                    true,
		"tag:winter":                    false,
		"[type:plant]":                  true,
		"[type:animal]":                 false,
		"[type]":                        true,
		"[type:animal] OR [type:plant]": true,
		"path:Garden tag:summer":        true,
		"path:Garden -tag:summer":       false,
		"wall":                          true,
		`path:"Projects/Garden"`:        true,
		"[type:animal] OR path:Kitchen": false,
	}

	for query, expected := range cases {
		if matchesQuery(query, page) != expected {
			t.Errorf("%s should be %v", query, expected)
		}
	}
}

func TestAddPageColorFirstGroupWins(t *testing.T) {
	colorGroups = []colorGroup{{Query: "path:Garden"}, {Query: "tag:summer"}}
	colorGroups[0].Color.RGB = 0x2e7d32
	colorGroups[1].Color.RGB = 0xff9800
	defer func() { colorGroups = nil; pageColors = map[string]string{} }()

	AddPageColor("garden/tomatoes", "Garden/Tomatoes.md", []byte("#summer"))
	AddPageColor("kitchen/salad", "Kitchen/Salad.md", []byte("#summer"))
	AddPageColor("kitchen/bread", "Kitchen/Bread.md", []byte("plain"))

	if pageColors["garden/tomatoes"] != "#2e7d32" || pageColors["kitchen/salad"] != "#ff9800" {
		t.Errorf("unexpected colours: %v", pageColors)
	}
	if _, colored := pageColors["kitchen/bread"]; colored {
		t.Errorf("a page without a group must keep the default colour")
	}
}
