package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/akrck02/papiro-indexer/logger"
	"github.com/akrck02/papiro-indexer/model"
)

const htmlExtension = "html"

func WriteIndex(index *model.IndexItem, path string) {

	json, error := json.Marshal(index)
	if nil != error {
		fmt.Println("Error converting to json format :", error.Error())
		return
	}

	error = os.WriteFile(fmt.Sprintf("%s/%s", path, "index.json"), json, defaultFilePermissions)
}

var obsidianVault *ObsidianVault

// PrepareObsidianVault loads the vault references when indexing an obsidian project.
func PrepareObsidianVault(configuration *model.IndexerConfiguration) {
	if configuration.IsObsidianProject {
		obsidianVault = LoadObsidianVault(configuration.Path)
		LoadColorGroups(configuration.Path)
	}
}

func IndexPath(configuration *model.IndexerConfiguration, parentItem *model.IndexItem, dirPath string) {

	// files not allowed
	if parentItem.Type == model.File {
		logger.Warning(parentItem.Path, "is a file.")
		return
	}

	// try to open the directory
	directory, error := OpenDirectory(dirPath)
	if nil != error {
		logger.Error(error.Error())
		return
	}
	defer directory.Close()

	// get children
	files, error := os.ReadDir(directory.Name())
	if nil != error {
		logger.Error("Cannot access the directory:", error.Error())
		return
	}

	// iterate over each item
	for _, file := range files {

		// obsidian configuration, trash and hidden files are not published
		if configuration.IsObsidianProject && IsObsidianIgnored(file.Name()) {
			continue
		}

		// check if file is accesible
		info, error := file.Info()
		if nil != error {
			logger.Warning("Error accessing file", dirPath, "/", info.Name(), "skipping...")
			break
		}

		// check if it is a directory or file, then index it (and children)
		// the indexed path and the original path
		if info.IsDir() {
			order, clean := SplitOrder(file.Name())
			name := EncodeUrl(clean)
			subitem := &model.IndexItem{
				Type:  model.Directory,
				Path:  name,
				Files: make(map[string]model.IndexItem),
				Order: order,
			}
			subDirPath := CreateUrl(dirPath, info.Name())
			IndexPath(configuration, subitem, subDirPath)

			if 0 != len(subitem.Files) {
				existing, exists := parentItem.Files[name]
				if exists {
					subitem.Path = existing.Path
					if nil == subitem.Order {
						subitem.Order = existing.Order
					}
				}
				parentItem.Files[name] = *subitem
			}

		} else {
			filePath := CreateUrl(dirPath, file.Name())
			indexFile(configuration, parentItem, filePath, file.Name())
		}
	}
}

func indexFile(configuration *model.IndexerConfiguration, parentItem *model.IndexItem, filePath string, name string) {
	extension := path.Ext(filePath)
	switch extension {
	case ".md":
		indexMarkdownFile(configuration, parentItem, filePath, name)
	case ".html":
		indexHtmlFile(configuration, parentItem, filePath, name)
	default:
		indexNonMarkupLanguagefile(configuration, parentItem, filePath)
	}
}

func indexMarkdownFile(configuration *model.IndexerConfiguration, parentItem *model.IndexItem, filePath string, name string) {

	order, cleanName := SplitOrder(name)
	newFileName := ChangeExtension(cleanName, htmlExtension)
	newFileUrl := CreateUrl(configuration.Destination, EncodeUrl(CleanPath(RemoveExtension(RemoveUrlStart(filePath, configuration.Path)), true)+"."+htmlExtension))

	subitem := &model.IndexItem{
		Type:  model.File,
		Path:  EncodeUrl(newFileName),
		Files: make(map[string]model.IndexItem),
		Order: order,
	}

	bytes, error := ReadFile(filePath)
	if nil != error {
		logger.Error(error.Error())
		return
	}

	relativePath := strings.TrimPrefix(RemoveUrlStart(filePath, configuration.Path), "/")
	route := EncodeUrl(CleanPath(RemoveExtension(relativePath), true))

	if nil != obsidianVault {
		AddPageColor(route, relativePath, bytes)
		bytes = obsidianVault.ObsidianToMarkdown(bytes)
	}

	// what search.json holds, from the markdown papiro shows (obsidian syntax converted)
	AddPageSearch(route, bytes)

	documentDir := path.Dir(strings.TrimPrefix(RemoveUrlStart(filePath, configuration.Path), "/"))
	bytes, links := MarkdownToHtml(bytes, documentDir)
	AddPageLinks(route, links)
	error = WriteFile(bytes, newFileUrl)
	if nil != error {
		logger.Error(error.Error())
		return
	}

	logger.Log("▸", "📜", path.Base(filePath), "⤳ ", newFileUrl)

	addToParent(parentItem, EncodeUrl(RemoveExtension(cleanName)), subitem)
}

func getNewFileRoute(configuration *model.IndexerConfiguration, filePath string, newExtension string) string {
	return fmt.Sprintf("%s.%s", RemoveExtension(RemoveUrlStart(filePath, configuration.Path)), newExtension)
}

func indexHtmlFile(configuration *model.IndexerConfiguration, parentItem *model.IndexItem, filePath string, name string) {

	order, cleanName := SplitOrder(name)
	newFileName := ChangeExtension(cleanName, htmlExtension)
	newFileUrl := CreateUrl(configuration.Destination, EncodeUrl(CleanPath(RemoveUrlStart(filePath, configuration.Path), true)))

	subitem := &model.IndexItem{
		Type:  model.File,
		Path:  EncodeUrl(newFileName),
		Files: make(map[string]model.IndexItem),
		Order: order,
	}

	bytes, error := ReadFile(filePath)
	if nil != error {
		logger.Error(error.Error())
		return
	}

	error = WriteFile(bytes, newFileUrl)
	if nil != error {
		logger.Error(error.Error())
		return
	}

	logger.Log("▸", "📜", path.Base(filePath), "⤳ ", newFileUrl)

	addToParent(parentItem, EncodeUrl(RemoveExtension(cleanName)), subitem)

}

func indexNonMarkupLanguagefile(configuration *model.IndexerConfiguration, _ *model.IndexItem, filePath string) {

	newRoute := fmt.Sprintf("%s%s", configuration.Destination, EncodeUrl(CleanPath(RemoveUrlStart(filePath, configuration.Path), false)))
	file, error := OpenFile(filePath)
	if nil != error {
		logger.Error(error.Error())
		return
	}

	CopyFile(file, newRoute)
	// logger.Log("▸", "🖼️ ", path.Base(filePath), "⤳ ", newRoute)
}

// addToParent adds a page to its folder; a page with the name of a folder
// next to it becomes the page of that folder, keeping its pages
func addToParent(parentItem *model.IndexItem, index string, subitem *model.IndexItem) {
	existing, exists := parentItem.Files[index]
	if !exists {
		parentItem.Files[index] = *subitem
		return
	}

	order := existing.Order
	if nil == order {
		order = subitem.Order
	}
	parentItem.Files[index] = model.IndexItem{
		Path:  subitem.Path,
		Type:  existing.Type,
		Files: existing.Files,
		Order: order,
	}
}
