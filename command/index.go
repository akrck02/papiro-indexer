package command

import (
	"fmt"
	"os"
	"time"

	"github.com/akrck02/papiro-indexer/model"
	"github.com/akrck02/papiro-indexer/service"
)

func Index(filePath string) {
	startTime := time.Now()
	index := model.IndexItem{
		Type:  model.Directory,
		Path:  "/",
		Files: make(map[string]model.IndexItem),
	}
	service.IndexPath(&index, filePath, filePath)
	service.WriteIndex(&index, os.Getenv("WIKI_PATH"))
	fmt.Println("\n", "⤷ Indexed in", time.Now().Sub(startTime).Milliseconds(), "ms.")
}
