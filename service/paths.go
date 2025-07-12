package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func EncodeUrl(url string) string {
	url = strings.TrimSpace(url)
	return strings.ToLower(strings.ReplaceAll(url, " ", "-"))
}

func RemovePathStart(path string, start string) string {
	newPath, _ := strings.CutPrefix(path, start)
	return newPath
}

func OpenDirectory(fileUrl string) (*os.File, error) {

	// If input file does not exists, raise an error
	fileStats, error := os.Stat(fileUrl)
	if os.IsNotExist(error) {
		return nil, errors.New("path does not exist.")
	}

	if nil != error {
		return nil, errors.New(fmt.Sprint("Cannot access stats for path", fileUrl, ":", error.Error()))
	}

	// if it is a file
	if !fileStats.IsDir() {
		return nil, errors.New("The given path is not a directory.")

	}

	// open the directory
	inputFile, error := os.Open(fileUrl)
	if nil != error {
		return nil, errors.New(fmt.Sprint("Cannot access the directory:", error.Error()))
	}

	return inputFile, nil
}

func OpenFile(fileUrl string) (*os.File, error) {

	// If input file does not exists, raise an error
	fileStats, error := os.Stat(fileUrl)
	if os.IsNotExist(error) {
		return nil, errors.New("path does not exist.")
	}

	if nil != error {
		return nil, errors.New(fmt.Sprint("Cannot access stats for path", fileUrl, ":", error.Error()))
	}

	// if it is a file
	if !fileStats.IsDir() {
		return nil, errors.New("The given path is not a file.")
	}

	// open the directory
	inputFile, error := os.Open(fileUrl)
	if nil != error {
		return nil, errors.New(fmt.Sprint("Cannot access the file:", error.Error()))
	}

	return inputFile, nil
}

func CopyFile(file io.Reader, url string) (int64, error) {

	destination, err := os.Create(url)
	if err != nil {
		return 0, err
	}
	defer destination.Close()

	return io.Copy(destination, file)
}

func RemoveExtension(url string) string {
	extension := filepath.Ext(url)
	name := url[0 : len(url)-len(extension)]
	return name
}
