package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/akrck02/papiro-indexer/logger"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// maxSearchText keeps search.json small: the text of a page beyond it is not searched
const maxSearchText = 20000

// SearchSection is a heading of a page, to open the page at it
type SearchSection struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// SearchPage is what the search of papiro looks in for a page
type SearchPage struct {
	Title    string          `json:"title"`
	Sections []SearchSection `json:"sections"`
	Text     string          `json:"text"`
}

var searchPages = map[string]SearchPage{}

// AddPageSearch keeps the title, sections and text of a page for search.json
func AddPageSearch(route string, md []byte) {
	p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock)
	doc := p.Parse(md)

	page := SearchPage{Sections: []SearchSection{}}
	var text strings.Builder

	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		switch item := node.(type) {
		case *ast.Heading:
			heading := strings.TrimSpace(plainText(item))
			if 1 == item.Level && "" == page.Title {
				page.Title = heading
			} else if 2 <= item.Level && item.Level <= 3 {
				page.Sections = append(page.Sections, SearchSection{ID: item.HeadingID, Text: heading})
			}
			text.WriteString(heading + "\n")
			return ast.SkipChildren
		case *ast.Text:
			text.Write(item.Literal)
			text.WriteString(" ")
		case *ast.Code:
			text.Write(item.Literal)
			text.WriteString(" ")
		case *ast.CodeBlock:
			text.Write(item.Literal)
			text.WriteString("\n")
		}
		return ast.GoToNext
	})

	page.Text = strings.Join(strings.Fields(text.String()), " ")
	if len(page.Text) > maxSearchText {
		page.Text = strings.ToValidUTF8(page.Text[:maxSearchText], "")
	}
	searchPages[route] = page
}

func plainText(node ast.Node) string {
	var text strings.Builder
	ast.WalkFunc(node, func(child ast.Node, entering bool) ast.WalkStatus {
		if entering {
			if leaf := child.AsLeaf(); nil != leaf {
				text.Write(leaf.Literal)
			}
		}
		return ast.GoToNext
	})
	return text.String()
}

// WriteSearch writes search.json: wiki route -> title, sections and text
func WriteSearch(destination string) {
	data, err := json.Marshal(searchPages)
	if nil != err {
		logger.Error("Error converting the search index to json:", err.Error())
		return
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s", destination, "search.json"), data, defaultFilePermissions)
	if nil != err {
		logger.Error("Error writing the search index:", err.Error())
	}
}
