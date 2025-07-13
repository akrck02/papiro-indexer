package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/akrck02/papiro-indexer/logger"
	"github.com/akrck02/papiro-indexer/model"
)

func WriteIndex(index *model.IndexItem, path string) {

	json, error := json.Marshal(index)
	if nil != error {
		fmt.Println("Error converting to json format :", error.Error())
		return
	}

	error = os.WriteFile(fmt.Sprintf("%s/%s", path, "/index.json"), json, 0644)
}

func IndexPath(parentItem *model.IndexItem, basePath string, dirPath string) {

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

		// check if file is accesible
		info, error := file.Info()
		if nil != error {
			logger.Warning("Error accessing file", dirPath, "/", info.Name(), "skipping...")
			break
		}

		// check if it is a directory or file, then index it (and children)
		// the indexed path and the original path
		if info.IsDir() {
			name := EncodeUrl(file.Name())
			subitem := &model.IndexItem{
				Type:  model.Directory,
				Path:  name,
				Files: make(map[string]model.IndexItem),
			}
			subDirPath := fmt.Sprintf("%s/%s", dirPath, info.Name())
			IndexPath(subitem, basePath, subDirPath)

			if 0 != len(subitem.Files) {
				parentItem.Files[name] = *subitem
			}

		} else {
			filePath := fmt.Sprintf("%s/%s", dirPath, file.Name())
			indexFile(parentItem, basePath, filePath, file.Name())
		}
	}
}

func indexFile(parentItem *model.IndexItem, basePath string, filePath string, name string) {
	extension := path.Ext(filePath)
	switch extension {
	case ".md":
		indexMarkdownFile(parentItem, basePath, filePath, name)
	case ".html":
		indexHtmlFile(parentItem, basePath, filePath, name)
	default:
		indexNonMarkupLanguagefile(parentItem, basePath, filePath)
	}
}

func indexMarkdownFile(parentItem *model.IndexItem, basePath string, filePath string, name string) {

	newRoute := EncodeUrl(fmt.Sprintf("%s%s.%s", os.Getenv("WIKI_PATH"), RemoveExtension(RemovePathStart(filePath, basePath)), "md"))

	subitem := &model.IndexItem{
		Type:  model.File,
		Path:  EncodeUrl(name),
		Files: make(map[string]model.IndexItem),
	}

	file, error := OpenFile(filePath)
	if nil != error {
		logger.Error(error.Error())
		return
	}

	CopyFile(file, newRoute)
	logger.Log("▸", "📜", path.Base(filePath), "⤳ ", newRoute)
	parentItem.Files[name] = *subitem
}

func indexHtmlFile(parentItem *model.IndexItem, basePath string, filePath string, name string) {
	newRoute := RemoveExtension(EncodeUrl(filePath))
	subitem := &model.IndexItem{
		Type:  model.File,
		Path:  EncodeUrl(name),
		Files: make(map[string]model.IndexItem),
	}
	logger.Log("▸", "📜", path.Base(filePath), "⤳ ", newRoute)
	parentItem.Files[name] = *subitem
}

func indexNonMarkupLanguagefile(_ *model.IndexItem, basePath string, filePath string) {

	newRoute := EncodeUrl(fmt.Sprintf("%s%s", os.Getenv("WIKI_PATH"), RemovePathStart(filePath, basePath)))
	file, error := OpenFile(filePath)
	if nil != error {
		logger.Error(error.Error())
		return
	}

	CopyFile(file, newRoute)
	logger.Log("▸", "🖼️ ", path.Base(filePath), "⤳ ", newRoute)
}
