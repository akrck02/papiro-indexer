package service

import (
	"strings"
	"testing"
)

func TestResolveWikiUrl(t *testing.T) {
	cases := []struct {
		destination string
		dir         string
		isLink      bool
		expected    string
	}{
		{"text.md", "components", true, "./components/text"},
		{"text.md#section", "components", true, "./components/text"},
		{"../README.md", "components", true, "./readme"},
		{"../images/components/My Button.png", "components", false, "./images/components/my-button.png"},
		{"./other page.md", ".", true, "./other-page"},
		{"/Notes/Other%20Note.md", "deep/dir", true, "./notes/other-note"},
		{"../../outside.md", "components", true, "../../outside.md"},
		{"https://example.com/a.md", "components", true, "https://example.com/a.md"},
		{"mailto:someone@example.com", "", true, "mailto:someone@example.com"},
		{"#section", "components", true, "#section"},
	}

	for _, c := range cases {
		if result := ResolveWikiUrl(c.destination, c.dir, c.isLink); result != c.expected {
			t.Errorf("%s from %s: got %s, expected %s", c.destination, c.dir, result, c.expected)
		}
	}
}

func TestMarkdownToHtmlResolvesRelativeUrls(t *testing.T) {
	html := string(MarkdownToHtml([]byte("[Text](text.md) ![x](../images/x.png)"), "components"))
	for _, expected := range []string{`href="./components/text"`, `src="./images/x.png"`} {
		if !strings.Contains(html, expected) {
			t.Errorf("%s not found in %s", expected, html)
		}
	}
}
