package service

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

	error := os.MkdirAll(filepath.Dir(fileUrl), defaultFilePermissions)
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

// orderRegex finds the position of a page or folder in its name: "1-intro",
// "01-get-started". One or two digits, so dates (2026-10-17) are not taken.
var orderRegex = regexp.MustCompile(`^(\d{1,2})[-_ ](.+)$`)

// SplitOrder returns the position written in a name and the name without it
func SplitOrder(name string) (int, string) {
	match := orderRegex.FindStringSubmatch(name)
	if nil == match {
		return 0, name
	}

	order, _ := strconv.Atoi(match[1])
	return order, match[2]
}

// CleanPath removes the position prefixes of the folders of a path, and of
// its last section when it is a page (other files keep their name)
func CleanPath(path string, isPage bool) string {
	sections := strings.Split(path, "/")
	for i, section := range sections {
		if i == len(sections)-1 && !isPage {
			continue
		}
		_, sections[i] = SplitOrder(section)
	}
	return strings.Join(sections, "/")
}
