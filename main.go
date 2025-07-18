package main

import (
	"flag"
	"fmt"

	"github.com/akrck02/papiro-indexer/command"
	"github.com/akrck02/papiro-indexer/model"
)

// Log app title to standard output
func logAppTitle() {
	println("-------------------------------------")
	println("   📜 Papiro indexer by akrck02 📜   ")
	println("-------------------------------------")
	println()
}

func main() {

	PathFlag := flag.String("path", "", "papiro indexer -path ./my/directory")
	outputPathFlag := flag.String("destination", "", "papiro indexer -path ./my/directory -destination ./new/directory")
	isObsidianProjectFlag := flag.Bool("obsidian", false, "papiro indexer -path ./my/directory -destination ./new/directory -obsidian")

	flag.Parse()

	// Load path if present, help otherwise.
	path := *PathFlag
	if "" == path {
		command.Help()
		return
	}

	// Load destination if present.
	destination := *outputPathFlag
	if "" == destination {
		command.Help()
		return
	}

	isObsidianProject := *isObsidianProjectFlag

	logAppTitle()

	configuration := model.IndexerConfiguration{
		Path:              path,
		Destination:       destination,
		IsObsidianProject: isObsidianProject,
	}

	fmt.Println("path:", path)
	fmt.Println("destination:", destination)
	fmt.Println("is obsidian:", isObsidianProject)

	// Index files.
	command.Index(&configuration)
}
