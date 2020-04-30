package utils

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)
var filePattern = regexp.MustCompile(`^(.*)/([A-Za-z\-_\.0-9]+)\.yaml$`)

func WalkPath(output string) (map[string]int, error) {
	fI, fE := os.Stat(output)
	if fE != nil {
		return nil, fE
	}
	files := make([]string, 0)
	if fI.IsDir() {
		err := filepath.Walk(output, func(path string, info os.FileInfo, walkError error) error {
			if strings.HasSuffix(path, ".yaml") {
				files = append(files, path)
			}
			return walkError
		})
		if err != nil {
			log.Println("error with directory listing", err)
			return nil, err
		}
	}

	folders := make(map[string]int)
	for _, f := range files {
		if filePattern.MatchString(f) {
			matches := filePattern.FindAllStringSubmatch(f, -1)
			folderName := matches[0][1]
			value, ok := folders[folderName]
			if !ok {
				value = 0
			}
			folders[folderName] = value + 1
		}
	}
	return folders, nil
}