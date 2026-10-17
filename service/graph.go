package service

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/akrck02/papiro-indexer/logger"
)

// graph keeps the links between pages: wiki route -> linked wiki routes
var graph = map[string][]string{}

// AddPageLinks registers the links found in a page
func AddPageLinks(route string, links []string) {
	graph[route] = links
}

// WriteGraph writes graph.json with the links between existing pages,
// without duplicates or links of a page to itself
func WriteGraph(destination string) {
	result := map[string][]string{}
	for route, links := range graph {
		seen := map[string]bool{}
		targets := []string{}
		for _, link := range links {
			_, exists := graph[link]
			if !exists || link == route || seen[link] {
				continue
			}
			seen[link] = true
			targets = append(targets, link)
		}
		sort.Strings(targets)
		result[route] = targets
	}

	data, err := json.Marshal(result)
	if nil != err {
		logger.Error("Error converting the graph to json:", err.Error())
		return
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s", destination, "graph.json"), data, defaultFilePermissions)
	if nil != err {
		logger.Error("Error writing the graph:", err.Error())
	}
}
