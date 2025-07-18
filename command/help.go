package command

import "flag"

func Help() {

	println("-------------------------------------")
	println("      📜 Papiro indexer help 📜      ")
	println("-------------------------------------")
	flag.Usage()
}
