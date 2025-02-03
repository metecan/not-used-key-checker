package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// scanFile checks if any key exists in the given file
func scanFile(filePath string, keys map[string]struct{}, foundKeys map[string]bool, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error opening file:", filePath, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())

		mu.Lock()
		for key := range keys {
			if strings.Contains(line, key) {
				foundKeys[key] = true
			}
		}
		mu.Unlock()
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("error reading file:", filePath, err)
	}
}

func main() {
	jsonFilePath := flag.String("json", "", "Path to the JSON file")
	projectDir := flag.String("dir", "", "Directory to scan")
	extensions := flag.String("ext", ".js,.ts,.tsx,.jsx,.json", "Comma-separated list of file extensions")
	excludeDirs := flag.String("exclude", "node_modules,.git,.next,dist,build", "Comma-separated list of directories to exclude")
	flag.Parse()

	if *jsonFilePath == "" || *projectDir == "" {
		fmt.Println("Usage: go run main.go -json=<json_file_path> -dir=<project_directory> -ext=<extensions> -exclude=<directories>")
		os.Exit(1)
	}

	// Read JSON file
	data, err := os.ReadFile(*jsonFilePath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	// Parse JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	keys := make(map[string]struct{})
	foundKeys := make(map[string]bool)
	for key := range jsonData {
		keys[strings.ToLower(key)] = struct{}{}
		foundKeys[strings.ToLower(key)] = false
	}

	fmt.Printf("✅ Loaded %d keys from JSON\n", len(keys))

	extList := strings.Split(*extensions, ",")
	excluded := strings.Split(*excludeDirs, ",")
	fileList := []string{}

	// Collect matching files
	err = filepath.WalkDir(*projectDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Println("error accessing:", path, err)
			return nil
		}

		// Check if the current path is in the excluded directories
		for _, ex := range excluded {
			if strings.Contains(path, "/"+ex+"/") {
				return nil
			}
		}

		if !d.IsDir() {
			for _, ext := range extList {
				if strings.HasSuffix(strings.ToLower(path), ext) {
					fileList = append(fileList, path)
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error scanning project directory:", err)
		return
	}

	fmt.Printf("🔍 Found %d files to scan\n", len(fileList))

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, file := range fileList {
		wg.Add(1)
		go scanFile(file, keys, foundKeys, &wg, &mu)
	}

	wg.Wait()

	// Collect missing keys
	missingKeys := []string{}
	for key, found := range foundKeys {
		if !found {
			missingKeys = append(missingKeys, key)
		}
	}

	fmt.Printf("\n❌ MISSING KEYS: %d\n", len(missingKeys))

	// Write missing keys to a file
	if len(missingKeys) > 0 {
		file, err := os.Create("missing_keys.txt")
		if err != nil {
			fmt.Println("Error creating missing_keys.txt:", err)
			return
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		for _, key := range missingKeys {
			writer.WriteString(key + "\n")
		}
		writer.Flush()

		fmt.Println("\n❌ MISSING KEYS saved to missing_keys.txt")
	} else {
		fmt.Println("\n✅ All keys were found in the project.")
	}
}
