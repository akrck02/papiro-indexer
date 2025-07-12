package main

import (
	"flag"

	"github.com/akrck02/papiro-indexer/command"
)

// Log app title to standard output
func logAppTitle() {
	println("-------------------------------------")
	println("   📜 Papiro indexer by akrck02 📜   ")
	println("-------------------------------------")
	println()
}

func main() {

	logAppTitle()

	envPathFlag := flag.String("e", "", "-e ./my/env/file")
	PathFlag := flag.String("d", "", "-d ./my/directory")
	flag.Parse()

	// Open help if help flag is present.
	if "" != *PathFlag {

		// Load env file configuration if present.
		if "" != *envPathFlag {
			command.LoadEnv(*envPathFlag)
		}

		// Index files.
		command.Index(*PathFlag)
		return
	}

	command.Help()
}
