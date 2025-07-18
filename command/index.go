package command

import (
	"fmt"
	"time"

	"github.com/akrck02/papiro-indexer/model"
	"github.com/akrck02/papiro-indexer/service"
)

func Index(configuration *model.IndexerConfiguration) {
	startTime := time.Now()
	index := model.IndexItem{
		Type:  model.Directory,
		Path:  "/",
		Files: make(map[string]model.IndexItem),
	}
	service.IndexPath(configuration, &index, configuration.Path)
	service.WriteIndex(&index, configuration.Destination)
	fmt.Println("\n", "⤷ Indexed in", time.Now().Sub(startTime).Milliseconds(), "ms.")
}
