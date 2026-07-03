package service

import (
	"fmt"
	"io"
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

func accessFileStats(fileUrl string) (os.FileInfo, error) {
	// Security: Prevent directory traversal attacks
	if strings.Contains(fileUrl, "..") {
		return nil, fmt.Errorf("invalid path: directory traversal not allowed")
	}

	fileStats, err := os.Stat(fileUrl)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", fileUrl)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot access stats for path %s: %w", fileUrl, err)
	}
	return fileStats, nil
}

func OpenDirectory(fileUrl string) (*os.File, error) {
	fileStats, err := accessFileStats(fileUrl)
	if err != nil {
		return nil, err
	}

	if !fileStats.IsDir() {
		return nil, fmt.Errorf("the given path is not a directory: %s", fileUrl)
	}

	return os.Open(fileUrl)
}

func OpenFile(fileUrl string) (*os.File, error) {
	fileStats, err := accessFileStats(fileUrl)
	if err != nil {
		return nil, err
	}

	if fileStats.IsDir() {
		return nil, fmt.Errorf("the given path is a directory, expected a file: %s", fileUrl)
	}

	return os.Open(fileUrl)
}

func ReadFile(fileUrl string) ([]byte, error) {
	fileStats, err := accessFileStats(fileUrl)
	if err != nil {
		return nil, err
	}

	if fileStats.IsDir() {
		return nil, fmt.Errorf("the given path is a directory, expected a file: %s", fileUrl)
	}

	return os.ReadFile(fileUrl)
}

func CopyFile(file io.Reader, fileUrl string) (int64, error) {
	err := os.MkdirAll(filepath.Dir(fileUrl), 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to create directory for %s: %w", fileUrl, err)
	}

	destination, err := os.Create(fileUrl)
	if err != nil {
		return 0, fmt.Errorf("failed to create file %s: %w", fileUrl, err)
	}
	defer destination.Close()

	return io.Copy(destination, file)
}

func WriteFile(bytes []byte, fileUrl string) error {
	err := os.MkdirAll(filepath.Dir(fileUrl), defaultFilePermissions)
	if err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", fileUrl, err)
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

// ExtractNumber extracts the leading number from a filename for sorting.
// Returns 0 if no number is found at the start.
func ExtractNumber(filename string) int {
	var num int
	for _, r := range filename {
		if r >= '0' && r <= '9' {
			num = num*10 + int(r-'0')
		} else {
			break
		}
	}
	return num
}

// SortByNumber sorts a slice of filenames based on their leading number.
func SortByNumber(files []string) []string {
	sorted := make([]string, len(files))
	copy(sorted, files)
	
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			numI := ExtractNumber(sorted[i])
			numJ := ExtractNumber(sorted[j])
			if numI > numJ {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}
