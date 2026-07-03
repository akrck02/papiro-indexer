package service

import (
	"strings"
	"testing"
)

func TestParseObsidianWikilinks(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`[[My Page]]`, `<a href="my-page">My Page</a>`},
		{`[[Target|Alias]]`, `<a href="target">Alias</a>`},
		{`No links here`, `No links here`},
	}

	for _, tt := range tests {
		result := string(ParseObsidianWikilinks([]byte(tt.input)))
		if result != tt.expected {
			t.Errorf("ParseObsidianWikilinks(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestMarkdownToHtmlWithWikilinks(t *testing.T) {
	input := `# Hello
Check out [[My Page]] and [[Target|Alias]].`
	
	result := string(MarkdownToHtml([]byte(input)))
	if !strings.Contains(result, `<a href="my-page">My Page</a>`) {
		t.Errorf("MarkdownToHtml did not convert wikilink correctly: %s", result)
	}
	if !strings.Contains(result, `<a href="target">Alias</a>`) {
		t.Errorf("MarkdownToHtml did not convert aliased wikilink correctly: %s", result)
	}
}
