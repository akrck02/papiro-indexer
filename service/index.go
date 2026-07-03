package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/akrck02/papiro/logger"
	"github.com/akrck02/papiro/model"
)

const htmlExtension = "html"

func WriteIndex(index *model.IndexItem, path string) {
	jsonData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		logger.Error(fmt.Errorf("error converting to json format: %w", err))
		return
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s", path, "index.json"), jsonData, defaultFilePermissions)
	if err != nil {
		logger.Error(fmt.Errorf("error writing index file: %w", err))
	} else {
		logger.Debug("Index written successfully to", fmt.Sprintf("%s/index.json", path))
	}
}

func IndexPath(configuration *model.IndexerConfiguration, parentItem *model.IndexItem, dirPath string) {
	if parentItem.Type == model.File {
		logger.Warning(parentItem.Path, "is a file.")
		return
	}

	directory, err := OpenDirectory(dirPath)
	if err != nil {
		logger.Error(fmt.Errorf("failed to open directory %s: %w", dirPath, err))
		return
	}
	defer directory.Close()

	files, err := directory.Readdirnames(0)
	if err != nil {
		logger.Error(fmt.Errorf("failed to read directory %s: %w", dirPath, err))
		return
	}

	// Sort files by number for deterministic ordering
	files = SortByNumber(files)

	for _, file := range files {
		fullPath := filepath.Join(dirPath, file)
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			logger.Warning(fmt.Errorf("could not stat %s: %w", fullPath, err))
			continue
		}

		if fileInfo.IsDir() {
			newItem := &model.IndexItem{
				Type:  model.Directory,
				Path:  EncodeUrl(file),
				Files: make(map[string]model.IndexItem),
			}
			parentItem.Files[file] = *newItem
			IndexPath(configuration, newItem, fullPath)
		} else {
			indexFile(configuration, parentItem, fullPath, file)
		}
	}
}

func indexFile(configuration *model.IndexerConfiguration, parentItem *model.IndexItem, filePath string, name string) {
	extension := filepath.Ext(filePath)
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
	newFileName := ChangeExtension(name, htmlExtension)
	newFileUrl := CreateEncodedUrl(configuration.Destination, RemoveUrlStart(filePath, configuration.Path))

	subitem := &model.IndexItem{
		Type:  model.File,
		Path:  EncodeUrl(newFileName),
		Files: make(map[string]model.IndexItem),
	}

	content, err := ReadFile(filePath)
	if err != nil {
		logger.Error(fmt.Errorf("failed to read markdown file %s: %w", filePath, err))
		return
	}

	htmlContent := MarkdownToHtml(content)
	err = os.WriteFile(newFileUrl, htmlContent, defaultFilePermissions)
	if err != nil {
		logger.Error(fmt.Errorf("failed to write html file %s: %w", newFileUrl, err))
		return
	}

	parentItem.Files[name] = *subitem
	logger.Debug("Indexed markdown:", filePath)
}

func getNewFileRoute(configuration *model.IndexerConfiguration, filePath string, newExtension string) string {
	return fmt.Sprintf("%s.%s", RemoveExtension(RemoveUrlStart(filePath, configuration.Path)), newExtension)
}

func indexHtmlFile(configuration *model.IndexerConfiguration, parentItem *model.IndexItem, filePath string, name string) {
	newFileName := ChangeExtension(name, htmlExtension)
	newFileUrl := CreateEncodedUrl(configuration.Destination, RemoveUrlStart(filePath, configuration.Path))

	subitem := &model.IndexItem{
		Type:  model.File,
		Path:  EncodeUrl(newFileName),
		Files: make(map[string]model.IndexItem),
	}

	content, err := ReadFile(filePath)
	if err != nil {
		logger.Error(fmt.Errorf("failed to read html file %s: %w", filePath, err))
		return
	}

	err = os.WriteFile(newFileUrl, content, defaultFilePermissions)
	if err != nil {
		logger.Error(fmt.Errorf("failed to write html file %s: %w", newFileUrl, err))
		return
	}

	parentItem.Files[name] = *subitem
	logger.Debug("Indexed html:", filePath)
}

func indexNonMarkupLanguagefile(configuration *model.IndexerConfiguration, _ *model.IndexItem, filePath string) {
	newRoute := EncodeUrl(fmt.Sprintf("%s%s", configuration.Destination, RemoveUrlStart(filePath, configuration.Path)))
	file, err := OpenFile(filePath)
	if err != nil {
		logger.Error(fmt.Errorf("failed to open file %s: %w", filePath, err))
		return
	}
	defer file.Close()

	_, err = CopyFile(file, newRoute)
	if err != nil {
		logger.Error(fmt.Errorf("failed to copy file %s: %w", filePath, err))
		return
	}
	logger.Debug("Copied asset:", filePath)
}
