package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/russross/blackfriday/v2"
)

// ParseObsidianWikilinks replaces [[WikiLink]] with HTML anchors.
func ParseObsidianWikilinks(md []byte) []byte {
	re := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	return re.ReplaceAllFunc(md, func(match []byte) []byte {
		link := match[2 : len(match)-2] // Extract content between [[ and ]]
		// Handle aliases: [[target|alias]] -> <a href="target">alias</a>
		parts := strings.Split(string(link), "|")
		target := parts[0]
		display := string(link)
		if len(parts) > 1 {
			display = parts[1]
		}
		
		// Convert target to URL-safe format (lowercase, hyphens for spaces)
		urlTarget := EncodeUrl(target)
		
		return []byte(fmt.Sprintf(`<a href="%s">%s</a>`, urlTarget, display))
	})
}

func MarkdownToHtml(md []byte) []byte {
	// First, parse Obsidian wikilinks if present
	md = ParseObsidianWikilinks(md)

	// create markdown parser with extensions
	extensions := blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs | blackfriday.NoEmptyLineBeforeBlockQuote
	p := blackfriday.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := blackfriday.CommonFlags | blackfriday.HrefTargetBlank
	opts := blackfriday.RendererOptions{Flags: htmlFlags}
	renderer := blackfriday.NewRenderer(opts)

	return p.Render(doc, renderer)
}
