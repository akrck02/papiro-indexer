package service

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/gomarkdown/markdown/ast"
)

// highlightCodeBlocks renders fenced code blocks with syntax highlighting:
// each token is a span with the short chroma class (k keyword, s string,
// c comment...), papiro colours them with its theme.
func highlightCodeBlocks(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	block, ok := node.(*ast.CodeBlock)
	if !ok {
		return ast.GoToNext, false
	}

	language := ""
	if fields := strings.Fields(string(block.Info)); 0 < len(fields) {
		language = strings.ToLower(fields[0])
	}

	class := ""
	if "" != language {
		class = fmt.Sprintf(` class="language-%s"`, html.EscapeString(language))
	}

	fmt.Fprintf(w, "<pre><code%s>", class)
	writeHighlighted(w, language, string(block.Literal))
	io.WriteString(w, "</code></pre>\n")
	return ast.GoToNext, true
}

func writeHighlighted(w io.Writer, language string, code string) {
	var lexer chroma.Lexer
	if "" != language {
		lexer = lexers.Get(language)
	}

	// unknown or no language: plain text
	if nil == lexer {
		io.WriteString(w, html.EscapeString(code))
		return
	}

	iterator, err := chroma.Coalesce(lexer).Tokenise(nil, code)
	if nil != err {
		io.WriteString(w, html.EscapeString(code))
		return
	}

	for _, token := range iterator.Tokens() {
		text := html.EscapeString(token.Value)
		class := chroma.StandardTypes[token.Type]
		if "" == class || chroma.Text == token.Type || chroma.TextWhitespace == token.Type {
			io.WriteString(w, text)
			continue
		}
		fmt.Fprintf(w, `<span class="%s">%s</span>`, class, text)
	}
}
