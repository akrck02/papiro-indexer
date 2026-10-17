package service

import (
	"fmt"
	"io/fs"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// ObsidianVault resolves obsidian style references ([[note]], ![[image.png]])
// to the paths papiro serves, relative to the vault root.
type ObsidianVault struct {
	byName map[string]string // lowercase file name (with and without .md) -> vault relative path
	byPath map[string]string // lowercase vault relative path (with and without .md) -> vault relative path
}

var (
	embedRegex     = regexp.MustCompile(`!\[\[([^\]\n]+)\]\]`)
	wikiLinkRegex  = regexp.MustCompile(`\[\[([^\]\n]+)\]\]`)
	highlightRegex = regexp.MustCompile(`==([^=\n]+)==`)
	commentRegex   = regexp.MustCompile(`(?s)%%.*?%%`)
	calloutRegex   = regexp.MustCompile(`^((?:>\s*)+)\[!([A-Za-z-]+)\][+-]?\s*(.*)$`)
	tagRegex       = regexp.MustCompile(`(^|\s)#([\p{L}_][\p{L}\p{N}_/-]*)`)
	imageSizeRegex = regexp.MustCompile(`^(\d+)(?:x(\d+))?$`)
	inlineCodeRx   = regexp.MustCompile("`[^`\n]*`")
)

var imageExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".webp": true, ".svg": true, ".bmp": true, ".avif": true,
}

// IsObsidianIgnored tells if a directory or file must be skipped in an obsidian vault.
func IsObsidianIgnored(name string) bool {
	return strings.HasPrefix(name, ".")
}

// LoadObsidianVault walks the vault and registers every file for reference resolution.
func LoadObsidianVault(root string) *ObsidianVault {
	vault := &ObsidianVault{
		byName: make(map[string]string),
		byPath: make(map[string]string),
	}

	filepath.WalkDir(root, func(filePath string, entry fs.DirEntry, err error) error {
		if nil != err {
			return nil
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
			return nil
		}
		relative = filepath.ToSlash(relative)
		vault.register(relative)
		return nil
	})

	return vault
}

func (vault *ObsidianVault) register(relative string) {
	keys := []string{strings.ToLower(relative)}
	names := []string{strings.ToLower(filepath.Base(relative))}
	if ".md" == strings.ToLower(filepath.Ext(relative)) {
		keys = append(keys, strings.ToLower(RemoveExtension(relative)))
		names = append(names, strings.ToLower(RemoveExtension(filepath.Base(relative))))
	}

	for _, key := range keys {
		vault.byPath[key] = relative
	}

	// obsidian resolves names to the shortest path when there are duplicates
	for _, name := range names {
		current, exists := vault.byName[name]
		if !exists || strings.Count(relative, "/") < strings.Count(current, "/") {
			vault.byName[name] = relative
		}
	}
}

// resolve returns the vault relative path of the target, or "" if not found.
func (vault *ObsidianVault) resolve(target string) string {
	target = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(target), "/"))
	if found, ok := vault.byPath[target]; ok {
		return found
	}

	if found, ok := vault.byName[strings.ToLower(filepath.Base(target))]; ok {
		return found
	}

	return ""
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

// ObsidianToMarkdown converts obsidian flavoured markdown to standard markdown.
func (vault *ObsidianVault) ObsidianToMarkdown(md []byte) []byte {
	text := strings.ReplaceAll(string(md), "\r\n", "\n")
	text = removeFrontmatter(text)
	text = commentRegex.ReplaceAllString(text, "")

	lines := strings.Split(text, "\n")
	inCode := false
	fence := ""
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			marker := trimmed[:3]
			if !inCode {
				inCode, fence = true, marker
			} else if marker == fence {
				inCode = false
			}
			continue
		}

		if inCode {
			continue
		}

		lines[i] = vault.convertLine(line)
	}

	return []byte(strings.Join(lines, "\n"))
}

func (vault *ObsidianVault) convertLine(line string) string {
	if match := calloutRegex.FindStringSubmatch(line); nil != match {
		title := strings.TrimSpace(match[3])
		if "" == title {
			title = strings.ToUpper(match[2][:1]) + strings.ToLower(match[2][1:])
		}
		line = fmt.Sprintf("%s**%s**", match[1], title)
	}

	// keep inline code untouched
	codes := inlineCodeRx.FindAllString(line, -1)
	line = inlineCodeRx.ReplaceAllString(line, "\x00")

	line = embedRegex.ReplaceAllStringFunc(line, vault.convertEmbed)
	line = wikiLinkRegex.ReplaceAllStringFunc(line, vault.convertWikiLink)
	line = highlightRegex.ReplaceAllString(line, "<mark>$1</mark>")
	line = tagRegex.ReplaceAllString(line, "$1`#$2`")

	for _, code := range codes {
		line = strings.Replace(line, "\x00", code, 1)
	}

	return line
}

// splitReference splits "target#section|alias" in its parts.
func splitReference(reference string) (target string, section string, alias string) {
	target, alias, _ = strings.Cut(reference, "|")
	target, section, _ = strings.Cut(target, "#")
	return strings.TrimSpace(target), strings.TrimSpace(section), strings.TrimSpace(alias)
}

func (vault *ObsidianVault) convertEmbed(match string) string {
	target, section, alias := splitReference(match[3 : len(match)-2])
	resolved := vault.resolve(target)
	if "" == resolved {
		return target
	}

	if imageExtensions[strings.ToLower(filepath.Ext(resolved))] {
		attributes := ""
		if size := imageSizeRegex.FindStringSubmatch(alias); nil != size {
			attributes = fmt.Sprintf(` width="%s"`, size[1])
			if "" != size[2] {
				attributes += fmt.Sprintf(` height="%s"`, size[2])
			}
			alias = ""
		}
		if "" == alias {
			alias = RemoveExtension(filepath.Base(resolved))
		}
		if "" == attributes {
			return fmt.Sprintf("![%s](%s)", alias, rootPath(resolved))
		}
		return fmt.Sprintf(`<img src="%s" alt="%s"%s>`, webPath(resolved), alias, attributes)
	}

	// note and other file embeds are shown as links
	return vault.link(resolved, target, section, alias)
}

func (vault *ObsidianVault) convertWikiLink(match string) string {
	target, section, alias := splitReference(match[2 : len(match)-2])

	// links to sections of the same note
	if "" == target {
		if "" == alias {
			return section
		}
		return alias
	}

	resolved := vault.resolve(target)
	if "" == resolved {
		if "" != alias {
			return alias
		}
		return target
	}

	return vault.link(resolved, target, section, alias)
}

func (vault *ObsidianVault) link(resolved string, target string, section string, alias string) string {
	text := alias
	if "" == text {
		text = filepath.Base(target)
		if "" != section {
			text += " > " + section
		}
	}

	text = strings.NewReplacer("[", `\[`, "]", `\]`).Replace(text)
	return fmt.Sprintf("[%s](%s)", text, rootPath(resolved))
}

func removeFrontmatter(text string) string {
	if !strings.HasPrefix(text, "---\n") {
		return text
	}

	end := strings.Index(text[4:], "\n---")
	if -1 == end {
		return text
	}

	rest := text[4+end+4:]
	newLine := strings.Index(rest, "\n")
	if -1 == newLine {
		return ""
	}

	return rest[newLine+1:]
}
