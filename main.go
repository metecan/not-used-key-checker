package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	jsonFilePath := flag.String("json", "", "path to the json file")
	projectDir := flag.String("dir", "", "directory to scan for usage")
	extensions := flag.String("ext", ".js,.ts,.tsx", "comma-separated list of file extensions to check")
	flag.Parse()

	if *jsonFilePath == "" || *projectDir == "" {
		fmt.Println("usage: go run main.go -json=<json_file_path> -dir=<project_directory> -ext=<extensions>")
		os.Exit(1)
	}

	data, err := os.ReadFile(*jsonFilePath)
	if err != nil {
		fmt.Println("failed to read the json file:", err)
		return
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		fmt.Println("failed to parse json:", err)
		return
	}

	extList := strings.Split(*extensions, ",")
	var fileList []string

	err = filepath.WalkDir(*projectDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			for _, ext := range extList {
				if strings.HasSuffix(path, ext) {
					fileList = append(fileList, path)
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("failed to scan project directory:", err)
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	notFoundKeys := make(map[string]bool)

	for key := range jsonData {
		notFoundKeys[key] = true
	}

	for _, file := range fileList {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			content, err := os.ReadFile(file)
			if err != nil {
				fmt.Println("failed to read file:", file, err)
				return
			}

			mu.Lock()
			for key := range jsonData {
				if strings.Contains(string(content), key) {
					delete(notFoundKeys, key)
				}
			}
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	for key := range notFoundKeys {
		fmt.Println("not found:", key)
	}
}
