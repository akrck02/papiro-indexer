package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/akrck02/papiro-indexer/logger"
	"github.com/akrck02/papiro-indexer/model"
)

func WriteIndex(index *map[string]model.IndexItem, path string) {

	json, error := json.Marshal(index)
	if nil != error {
		fmt.Println("Error converting to json format :", error.Error())
		return
	}

	error = os.WriteFile(fmt.Sprintf("%s/%s", path, "/index.json"), json, 0644)
}

func IndexPath(index map[string]model.IndexItem, dirPath string) {

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

	// get relative path

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
		name := EncodeUrl(file.Name())
		subitem := &model.IndexItem{
			Path: name,
		}

		if info.IsDir() {
			subitem.Type = model.Directory
			subitem.Files = make(map[string]model.IndexItem)
			indexChildren(subitem, dirPath)
		} else {
			subitem.Type = model.File
		}
		index[name] = *subitem
	}

}

func indexChildren(parentItem *model.IndexItem, dirPath string) {

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
		name := EncodeUrl(file.Name())

		if info.IsDir() {
			subitem := &model.IndexItem{
				Type:  model.Directory,
				Path:  name,
				Files: make(map[string]model.IndexItem),
			}
			indexChildren(subitem, fmt.Sprintf("%s/%s", dirPath, info.Name()))
			parentItem.Files[name] = *subitem
		} else {
			logger.Log(dirPath)
			indexFile(parentItem, fmt.Sprintf("%s", dirPath), file.Name())
		}
	}
}

func indexFile(parentItem *model.IndexItem, filePath string, name string) {
	extension := path.Ext(filePath)
	switch extension {
	case ".md":
		indexMarkdownFile(parentItem, filePath, name)
	case ".html":
		indexHtmlFile(parentItem, filePath, name)
	default:
		indexNonMarkupLanguagefile(parentItem, filePath, name)
	}
}

func indexMarkdownFile(parentItem *model.IndexItem, filePath string, name string) {
	newRoute := os.Getenv("WIKI_PATH")
	newRoute += "/" + name
	newRoute = EncodeUrl(RemoveExtension(newRoute))

	logger.Log("▸", "📜", path.Base(filePath), "⤳ ", newRoute)
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

func indexHtmlFile(parentItem *model.IndexItem, filePath string, name string) {
	newRoute := RemoveExtension(EncodeUrl(filePath))
	subitem := &model.IndexItem{
		Type:  model.File,
		Path:  EncodeUrl(name),
		Files: make(map[string]model.IndexItem),
	}
	logger.Log("▸", "📜", path.Base(filePath), "⤳ ", newRoute)
	parentItem.Files[name] = *subitem
}

func indexNonMarkupLanguagefile(parentItem *model.IndexItem, path string, name string) {
	//logger.Log("▸", "🖼️", path.Base(filePath), "⤳ ", newRoute)
}
