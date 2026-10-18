package service

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// MarkdownToHtml renders the markdown of a document placed in documentDir
// (relative to the wiki root) to html. It also returns the wiki routes the
// document links to, which may not exist.
func MarkdownToHtml(md []byte, documentDir string) ([]byte, []string) {
	links := []string{}
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
			resolved := ResolveWikiUrl(string(item.Destination), documentDir, true)
			if route, ok := pageRoute(resolved); ok {
				links = append(links, route)
			}
			item.Destination = []byte(resolved)
		case *ast.Image:
			item.Destination = []byte(ResolveWikiUrl(string(item.Destination), documentDir, false))
		case *ast.HTMLBlock:
			item.Literal = resolveHtmlUrls(item.Literal, documentDir)
		case *ast.HTMLSpan:
			item.Literal = resolveHtmlUrls(item.Literal, documentDir)
		}
		return ast.GoToNext
	})

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags, RenderNodeHook: highlightCodeBlocks}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer), links
}

// pageRoute returns the wiki route of a resolved link when it can point to a page
func pageRoute(resolved string) (string, bool) {
	if !strings.HasPrefix(resolved, "./") {
		return "", false
	}

	route, err := url.PathUnescape(strings.TrimPrefix(resolved, "./"))
	if nil != err || "" == route || "" != path.Ext(route) {
		return "", false
	}

	return route, true
}

// htmlUrlRegex finds the addresses of html written in the markdown:
// <img src>, <source srcset>, <a href>…
var htmlUrlRegex = regexp.MustCompile(`(?i)\b(src|href|srcset)\s*=\s*("[^"]*"|'[^']*')`)

// resolveHtmlUrls points the relative addresses of html written in the
// markdown to the wiki root, as markdown links and images are
func resolveHtmlUrls(html []byte, documentDir string) []byte {
	return htmlUrlRegex.ReplaceAllFunc(html, func(match []byte) []byte {
		parts := htmlUrlRegex.FindSubmatch(match)
		attribute := strings.ToLower(string(parts[1]))
		quote := parts[2][:1]
		value := string(parts[2][1 : len(parts[2])-1])

		if "srcset" == attribute {
			// "a.png 1x, b.png 2x": each candidate is an address and a descriptor
			candidates := strings.Split(value, ",")
			for i, candidate := range candidates {
				fields := strings.Fields(candidate)
				if 0 == len(fields) {
					continue
				}
				fields[0] = ResolveWikiUrl(fields[0], documentDir, false)
				candidates[i] = strings.Join(fields, " ")
			}
			value = strings.Join(candidates, ", ")
		} else {
			value = ResolveWikiUrl(value, documentDir, "href" == attribute)
		}

		return []byte(fmt.Sprintf("%s=%s%s%s", parts[1], quote, value, quote))
	})
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

	// folders and pages are published without their "01-" position prefix
	isPage := isLink && ("" == path.Ext(target) || ".md" == strings.ToLower(path.Ext(target)))
	target = CleanPath(target, isPage)

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
