package service

import (
	"fmt"
	"html"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	obsidian "github.com/akrck02/go-markdown"
	obsidianvault "github.com/akrck02/go-markdown/vault"
)

// ObsidianVault resolves obsidian style references ([[note]], ![[image.png]]) to the paths papiro
// serves, relative to the vault root. Go Markdown reads the notes and resolves the links, so a link
// leads to the same note here as in Obsidian and in Remarkdown.
type ObsidianVault struct {
	files *obsidianvault.Vault
}

var imageExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".webp": true, ".svg": true, ".bmp": true, ".avif": true,
}

const (
	noteExtension     = ".md"
	sectionSeparator  = "#"
	sectionShownAfter = " > "
	strongMark        = "**"
	calloutOpening    = "[!"
)

// IsObsidianIgnored tells if a directory or file must be skipped in an obsidian vault.
func IsObsidianIgnored(name string) bool {
	return strings.HasPrefix(name, ".")
}

// LoadObsidianVault walks the vault and registers every file, and the aliases of every note, for
// reference resolution.
func LoadObsidianVault(root string) *ObsidianVault {
	var paths []string
	filepath.WalkDir(root, func(filePath string, entry fs.DirEntry, err error) error {
		if nil != err {
			return nil // an unreadable file is left out of the vault, as before
		}
		if filePath != root && IsObsidianIgnored(entry.Name()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, filePath)
		if nil != err {
			return nil // cannot happen: every path walked is inside root
		}
		paths = append(paths, filepath.ToSlash(relative))
		return nil
	})

	vault := newObsidianVault(paths)
	for _, relative := range paths {
		vault.addAliasesOf(root, relative)
	}
	return vault
}

func newObsidianVault(paths []string) *ObsidianVault {
	return &ObsidianVault{files: obsidianvault.New(paths)}
}

// addAliasesOf reads the aliases property of a note, so that a link may name it by one of them.
func (vault *ObsidianVault) addAliasesOf(root string, relative string) {
	if noteExtension != strings.ToLower(filepath.Ext(relative)) {
		return
	}
	text, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
	if nil != err {
		return // the indexer reports an unreadable note when it converts it
	}
	blocks := obsidian.Parse(string(text))
	if 0 < len(blocks) && obsidian.FrontMatter == blocks[0].Kind {
		vault.files.AddAliases(relative, blocks[0].Properties().Aliases())
	}
}

// webPath converts a vault relative path to the url papiro serves it from.
func webPath(relative string) string {
	return ResolveWikiUrl("/"+relative, "", true)
}

// rootPath converts a vault relative path to a markdown destination from the vault root.
func rootPath(relative string) string {
	sections := strings.Split(relative, "/")
	for i, section := range sections {
		sections[i] = url.PathEscape(section)
	}
	return "/" + strings.Join(sections, "/")
}

// ObsidianToMarkdown converts obsidian flavoured markdown to standard markdown. notePath is the
// note's path from the vault root, which links are resolved from. Only what Obsidian adds is
// rewritten; every other byte of the note is kept as it was.
func (vault *ObsidianVault) ObsidianToMarkdown(md []byte, notePath string) []byte {
	note := newSourceNote(strings.ReplaceAll(string(md), "\r\n", "\n"))
	var edits []edit
	for _, block := range obsidian.Parse(note.text) {
		switch block.Kind {
		case obsidian.FrontMatter, obsidian.Comment:
			edits = append(edits, note.removeLines(block.FirstLine, block.LastLine))
			continue
		case obsidian.Code:
			continue
		case obsidian.Callout:
			edits = append(edits, note.calloutEdits(block)...)
		}
		edits = append(edits, vault.spanEdits(note, block, filepath.ToSlash(notePath))...)
	}
	return []byte(note.applied(edits))
}

// sourceNote is a note's text and where each of its lines starts, to place edits in it.
type sourceNote struct {
	text       string
	lineStarts []int
}

func newSourceNote(text string) sourceNote {
	note := sourceNote{text: text, lineStarts: []int{0}}
	for index, character := range text {
		if '\n' == character {
			note.lineStarts = append(note.lineStarts, index+1)
		}
	}
	return note
}

// offset is the byte of the note at a line and column.
func (note sourceNote) offset(line int, column int) int {
	return note.lineStarts[line] + column
}

// offsetIn is the byte of the note at a byte of a block's text.
func (note sourceNote) offsetIn(block obsidian.Block, byteOfText int) int {
	return note.offset(block.SourceOf(byteOfText))
}

// edit replaces the bytes of the note from start to end.
type edit struct {
	start, end int
	text       string
}

// removeLines is the edit that takes lines out of the note, with their line breaks.
func (note sourceNote) removeLines(first int, last int) edit {
	end := len(note.text)
	if last+1 < len(note.lineStarts) {
		end = note.lineStarts[last+1]
	}
	return edit{start: note.lineStarts[first], end: end}
}

// applied is the note with the edits made, from the last to the first so that each keeps its place.
// An edit that overlaps one already made is left out, so that odd notes cannot break the page.
func (note sourceNote) applied(edits []edit) string {
	sort.SliceStable(edits, func(first, second int) bool {
		return edits[first].start > edits[second].start
	})
	text := note.text
	madeFrom := len(text) + 1
	for _, change := range edits {
		if change.end > madeFrom {
			continue
		}
		text = text[:change.start] + change.text + text[change.end:]
		madeFrom = change.start
	}
	return text
}

// spanEdits is the edits that turn the obsidian spans of a block, and those inside them, into
// standard markdown.
func (vault *ObsidianVault) spanEdits(note sourceNote, block obsidian.Block, notePath string) []edit {
	var edits []edit
	for _, span := range block.AllSpans() {
		start, end := note.offsetIn(block, span.Start), note.offsetIn(block, span.End)
		textStart, textEnd := note.offsetIn(block, span.TextStart), note.offsetIn(block, span.TextStart+len(span.Text))
		crossesLines := strings.Contains(note.text[start:end], "\n")
		switch {
		case obsidian.Commented == span.Style:
			edits = append(edits, edit{start: start, end: end}) // a hidden comment goes, even across lines
		case obsidian.Highlighted == span.Style:
			edits = append(edits, edit{start: start, end: textStart, text: "<mark>"}, edit{start: textEnd, end: end, text: "</mark>"})
		case crossesLines:
			continue // a link across lines is left as written, rather than losing its line break
		case obsidian.Tag == span.Link.Kind:
			edits = append(edits, edit{start: start, end: end, text: "`" + span.Text + "`"})
		case obsidian.WikiLink == span.Link.Kind:
			edits = append(edits, edit{start: start, end: end, text: vault.convertWikiLink(span, notePath)})
		case obsidian.WikiEmbed == span.Link.Kind:
			edits = append(edits, edit{start: start, end: end, text: vault.convertEmbed(span, notePath)})
		}
	}
	return edits
}

// calloutEdits turns "> [!warning] Careful" into "> **Careful**", and a callout without a title into
// its type, capitalised. Quote marks before the callout stay, so nested callouts keep their place.
func (note sourceNote) calloutEdits(block obsidian.Block) []edit {
	first := block.Pieces[0]
	written := block.Text[first.Start : first.Start+first.Length]
	lineStart := note.offset(first.Line, first.Column)
	markStart := lineStart + strings.Index(written, calloutOpening)
	lineEnd := lineStart + len(written)
	title := block.Header.Title
	if "" == title {
		return []edit{{start: markStart, end: lineEnd, text: strongMark + capitalised(block.Header.Type) + strongMark}}
	}
	titleStart := lineStart + strings.LastIndex(written, title)
	return []edit{
		{start: markStart, end: titleStart, text: strongMark},
		{start: titleStart + len(title), end: titleStart + len(title), text: strongMark},
	}
}

// capitalised is a word with its first letter in capitals.
func capitalised(word string) string {
	first, size := utf8.DecodeRuneInString(word)
	return string(unicode.ToUpper(first)) + word[size:]
}

// aliasOf is the text a wiki link shows instead of its target, or "" when it shows the target.
func aliasOf(span obsidian.Span) string {
	written := span.Link.Target
	if "" != span.Link.Section {
		written += sectionSeparator + span.Link.Section
	}
	if span.Text == written {
		return ""
	}
	return span.Text
}

func (vault *ObsidianVault) convertEmbed(span obsidian.Span, notePath string) string {
	link := span.Link
	resolved, resolution := vault.files.Resolve(notePath, link)
	if obsidianvault.NotFound == resolution {
		return link.Target
	}
	alias := aliasOf(span)
	if !imageExtensions[strings.ToLower(filepath.Ext(resolved))] {
		return vault.link(resolved, link.Target, link.Section, alias) // note and other file embeds are shown as links
	}
	if "" == alias {
		alias = RemoveExtension(filepath.Base(resolved))
	}
	if 0 == link.Width {
		return fmt.Sprintf("![%s](%s)", alias, rootPath(resolved))
	}
	attributes := fmt.Sprintf(` width="%d"`, link.Width)
	if 0 != link.Height {
		attributes += fmt.Sprintf(` height="%d"`, link.Height)
	}
	return fmt.Sprintf(`<img src="%s" alt="%s"%s>`, webPath(resolved), html.EscapeString(alias), attributes)
}

func (vault *ObsidianVault) convertWikiLink(span obsidian.Span, notePath string) string {
	link := span.Link
	alias := aliasOf(span)

	// links to sections of the same note
	if "" == link.Target {
		if "" == alias {
			return link.Section
		}
		return alias
	}

	resolved, resolution := vault.files.Resolve(notePath, link)
	if obsidianvault.NotFound == resolution {
		if "" != alias {
			return alias
		}
		return link.Target
	}

	return vault.link(resolved, link.Target, link.Section, alias)
}

func (vault *ObsidianVault) link(resolved string, target string, section string, alias string) string {
	text := alias
	if "" == text {
		text = filepath.Base(target)
		if "" != section {
			text += sectionShownAfter + section
		}
	}

	text = strings.NewReplacer("[", `\[`, "]", `\]`).Replace(text)
	return fmt.Sprintf("[%s](%s)", text, rootPath(resolved))
}
