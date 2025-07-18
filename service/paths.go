package service

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const defaultFilePermissions = 0766

func EncodeUrl(url string) string {
	url = strings.TrimSpace(url)
	return strings.ToLower(strings.ReplaceAll(url, " ", "-"))
}

func RemoveUrlStart(url string, start string) string {
	newPath, _ := strings.CutPrefix(url, start)
	return newPath
}

func accessFileStats(fileUrl string) (fs.FileInfo, error) {

	// If input file does not exists, raise an error
	fileStats, error := os.Stat(fileUrl)
	if os.IsNotExist(error) {
		return nil, errors.New("path does not exist.")
	}

	if nil != error {
		return nil, errors.New(fmt.Sprint("Cannot access stats for path", fileUrl, ":", error.Error()))
	}

	return fileStats, nil
}

func OpenDirectory(fileUrl string) (*os.File, error) {

	fileStats, error := accessFileStats(fileUrl)
	if nil != error {
		return nil, error
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
	fileStats, error := accessFileStats(fileUrl)
	if nil != error {
		return nil, error
	}

	// if it is a file
	if fileStats.IsDir() {
		return nil, errors.New("The given path is not a file.")
	}

	// open the directory
	inputFile, error := os.Open(fileUrl)
	if nil != error {
		return nil, errors.New(fmt.Sprint("Cannot access the file:", error.Error()))
	}

	return inputFile, nil
}

func ReadFile(fileUrl string) ([]byte, error) {

	// If input file does not exists, raise an error
	fileStats, error := accessFileStats(fileUrl)
	if nil != error {
		return nil, error
	}

	// if it is a file
	if fileStats.IsDir() {
		return nil, errors.New("The given path is not a file.")
	}

	// open the directory
	bytes, error := os.ReadFile(fileUrl)
	if nil != error {
		return nil, errors.New(fmt.Sprint("Cannot access the file:", error.Error()))
	}

	return bytes, nil
}

func CopyFile(file io.Reader, fileUrl string) (int64, error) {

	error := os.MkdirAll(filepath.Dir(fileUrl), 0644)
	if nil != error {
		return 0, error
	}

	destination, error := os.Create(fileUrl)
	if error != nil {
		return 0, error
	}
	defer destination.Close()

	return io.Copy(destination, file)
}

func WriteFile(bytes []byte, fileUrl string) error {

	error := os.MkdirAll(filepath.Dir(fileUrl), defaultFilePermissions)
	if nil != error {
		return error
	}

	return os.WriteFile(fileUrl, bytes, defaultFilePermissions)
}

func RemoveExtension(url string) string {
	extension := filepath.Ext(url)
	name := url[0 : len(url)-len(extension)]
	return name
}

func CreateEncodedUrl(sections ...string) string {
	return EncodeUrl(strings.Join(sections, "/"))
}

func CreateUrl(sections ...string) string {
	return strings.Join(sections, "/")
}

func RemoveSlashes(url string) string {
	return strings.ReplaceAll(url, "/", "")
}

func ChangeExtension(url string, extension string) string {
	return fmt.Sprintf("%s.%s", RemoveExtension(url), extension)
}
