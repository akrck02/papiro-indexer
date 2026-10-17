package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/akrck02/papiro-indexer/logger"
	"gopkg.in/yaml.v3"
)

// colorGroup is a colour rule of the obsidian graph: the pages matching
// the search query take the colour, the first matching group wins.
type colorGroup struct {
	Query string `json:"query"`
	Color struct {
		RGB int `json:"rgb"`
	} `json:"color"`
}

// colorPage is what a search query can look at
type colorPage struct {
	path       string // vault relative, with extension
	name       string // file name without extension
	content    string // the markdown, without frontmatter
	tags       []string
	properties map[string]any
}

var (
	colorGroups []colorGroup
	pageColors  = map[string]string{}
	termRegex   = regexp.MustCompile(`-?(?:\[[^\]]*\]|[a-z]+:"[^"]*"|"[^"]*"|\S+)`)
)

// LoadColorGroups reads the colour groups of the vault graph, if any
func LoadColorGroups(vaultPath string) {
	data, err := os.ReadFile(filepath.Join(vaultPath, ".obsidian", "graph.json"))
	if nil != err {
		return
	}

	var settings struct {
		ColorGroups []colorGroup `json:"colorGroups"`
	}
	if err := json.Unmarshal(data, &settings); nil != err {
		logger.Warning("Cannot read the graph colour groups:", err.Error())
		return
	}

	colorGroups = settings.ColorGroups
}

// AddPageColor finds the colour group of a page and keeps its colour
func AddPageColor(route string, relativePath string, md []byte) {
	if 0 == len(colorGroups) {
		return
	}

	page := newColorPage(relativePath, string(md))
	for _, group := range colorGroups {
		if matchesQuery(group.Query, page) {
			pageColors[route] = fmt.Sprintf("#%06x", group.Color.RGB&0xffffff)
			return
		}
	}
}

// WriteGraphColors writes graph-colors.json: wiki route -> colour
func WriteGraphColors(destination string) {
	if 0 == len(colorGroups) {
		return
	}

	data, err := json.Marshal(pageColors)
	if nil != err {
		logger.Error("Error converting the graph colours to json:", err.Error())
		return
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s", destination, "graph-colors.json"), data, defaultFilePermissions)
	if nil != err {
		logger.Error("Error writing the graph colours:", err.Error())
	}
}

func newColorPage(relativePath string, text string) colorPage {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	page := colorPage{
		path:       strings.TrimPrefix(filepath.ToSlash(relativePath), "/"),
		name:       RemoveExtension(filepath.Base(relativePath)),
		content:    removeFrontmatter(text),
		properties: map[string]any{},
	}

	if strings.HasPrefix(text, "---\n") {
		if end := strings.Index(text[4:], "\n---"); -1 != end {
			yaml.Unmarshal([]byte(text[4:4+end]), &page.properties)
		}
	}

	// tags from the frontmatter and from the text
	for key, value := range page.properties {
		if "tags" == strings.ToLower(key) || "tag" == strings.ToLower(key) {
			for _, tag := range propertyValues(value) {
				page.tags = append(page.tags, strings.TrimPrefix(tag, "#"))
			}
		}
	}
	for _, match := range tagRegex.FindAllStringSubmatch(page.content, -1) {
		page.tags = append(page.tags, match[2])
	}

	return page
}

// matchesQuery evaluates a search query of obsidian: alternatives joined by
// OR, each one a list of terms that must all match, "-" negates a term
func matchesQuery(query string, page colorPage) bool {
	for _, alternative := range strings.Split(query, " OR ") {
		terms := termRegex.FindAllString(strings.TrimSpace(alternative), -1)
		if 0 == len(terms) {
			continue
		}

		all := true
		for _, term := range terms {
			negated := strings.HasPrefix(term, "-") && 1 < len(term)
			if negated {
				term = term[1:]
			}
			if matchesTerm(term, page) == negated {
				all = false
				break
			}
		}

		if all {
			return true
		}
	}

	return false
}

func matchesTerm(term string, page colorPage) bool {
	contains := func(text string, value string) bool {
		return strings.Contains(strings.ToLower(text), strings.ToLower(value))
	}

	// [property] or [property:value]
	if strings.HasPrefix(term, "[") && strings.HasSuffix(term, "]") {
		key, value, hasValue := strings.Cut(term[1:len(term)-1], ":")
		for name, propertyValue := range page.properties {
			if !strings.EqualFold(name, strings.TrimSpace(key)) {
				continue
			}
			if !hasValue {
				return true
			}
			for _, item := range propertyValues(propertyValue) {
				if contains(item, unquote(strings.TrimSpace(value))) {
					return true
				}
			}
		}
		return false
	}

	operator, value, hasOperator := strings.Cut(term, ":")
	if hasOperator {
		value = unquote(value)
		switch strings.ToLower(operator) {
		case "path":
			return contains(page.path, value)
		case "file":
			return contains(page.name, value)
		case "tag":
			value = strings.ToLower(strings.TrimPrefix(value, "#"))
			for _, tag := range page.tags {
				tag = strings.ToLower(tag)
				if tag == value || strings.HasPrefix(tag, value+"/") {
					return true
				}
			}
			return false
		case "content":
			return contains(page.content, value)
		}
	}

	// plain text searches the name and the content
	value = unquote(term)
	return contains(page.name, value) || contains(page.content, value)
}

// propertyValues returns the values of a property as text
func propertyValues(value any) []string {
	switch typed := value.(type) {
	case []any:
		values := []string{}
		for _, item := range typed {
			values = append(values, fmt.Sprint(item))
		}
		return values
	case nil:
		return []string{}
	default:
		return []string{fmt.Sprint(typed)}
	}
}

func unquote(value string) string {
	if 2 <= len(value) && strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		return value[1 : len(value)-1]
	}
	return value
}
