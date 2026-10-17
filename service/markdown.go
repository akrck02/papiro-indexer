package service

import (
	"net/url"
	"path"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// MarkdownToHtml renders the markdown of a document placed in documentDir
// (relative to the wiki root) to html.
func MarkdownToHtml(md []byte, documentDir string) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// make relative links and images point to the wiki root
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		switch item := node.(type) {
		case *ast.Link:
			item.Destination = []byte(ResolveWikiUrl(string(item.Destination), documentDir, true))
		case *ast.Image:
			item.Destination = []byte(ResolveWikiUrl(string(item.Destination), documentDir, false))
		}
		return ast.GoToNext
	})

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

// ResolveWikiUrl converts a url relative to the document directory (or to the
// wiki root if it starts with "/") to a "./" url relative to the wiki root,
// which is how papiro resolves links and images.
// Links to markdown files point to their wiki page. Absolute urls, same page
// anchors and paths outside the wiki are left untouched.
func ResolveWikiUrl(destination string, documentDir string, isLink bool) string {
	if "" == destination || strings.HasPrefix(destination, "#") || strings.HasPrefix(destination, "//") {
		return destination
	}

	parsed, err := url.Parse(destination)
	if nil != err || "" != parsed.Scheme || "" != parsed.Host {
		return destination
	}

	target := parsed.Path
	if strings.HasPrefix(target, "/") {
		target = path.Clean(target[1:])
	} else {
		target = path.Clean(path.Join(documentDir, target))
	}

	if ".." == target || strings.HasPrefix(target, "../") {
		return destination
	}

	if "." == target {
		target = ""
	}

	// pages are served without extension, the section can't be kept because
	// papiro uses the url hash for routing
	if isLink && ".md" == strings.ToLower(path.Ext(target)) {
		target = RemoveExtension(target)
	}

	sections := strings.Split(EncodeUrl(target), "/")
	for i, section := range sections {
		sections[i] = url.PathEscape(section)
	}

	return "./" + strings.Join(sections, "/")
}
