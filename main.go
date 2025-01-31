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
func scanFile(filePath string, keys map[string]struct{}, foundKeys map[string]string, results chan<- string, wg *sync.WaitGroup, mu *sync.Mutex) {
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
		for key := range keys {
			if strings.Contains(line, key) {
				mu.Lock()
				foundKeys[key] = filePath
				mu.Unlock()
				results <- fmt.Sprintf("FOUND: %s in %s", key, filePath)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("error reading file:", filePath, err)
	}
}

func main() {
	jsonFilePath := flag.String("json", "", "Path to the JSON file")
	projectDir := flag.String("dir", "", "Directory to scan")
	extensions := flag.String("ext", ".js,.ts,.tsx,.jsx,.json", "Comma-separated list of file extensions")
	flag.Parse()

	if *jsonFilePath == "" || *projectDir == "" {
		fmt.Println("Usage: go run main.go -json=<json_file_path> -dir=<project_directory> -ext=<extensions>")
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
	for key := range jsonData {
		keys[strings.ToLower(key)] = struct{}{}
	}

	fmt.Printf("✅ Loaded %d keys from JSON\n", len(keys))

	extList := strings.Split(*extensions, ",")
	fileList := []string{}

	// Collect matching files
	err = filepath.WalkDir(*projectDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Println("error accessing:", path, err)
			return nil
		}
		if !d.IsDir() {
			for _, ext := range extList {
				if strings.HasSuffix(strings.ToLower(path), ext) {
					fileList = append(fileList, path)
					fmt.Println("Scanning file:", path)
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
	results := make(chan string, 100)
	foundKeys := make(map[string]string)
	var mu sync.Mutex

	for _, file := range fileList {
		wg.Add(1)
		go scanFile(file, keys, foundKeys, results, &wg, &mu)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	foundCount := 0
	for result := range results {
		fmt.Println(result)
		foundCount++
	}

	// List missing keys
	fmt.Println("\n📜 SUMMARY")
	fmt.Println("===========================")
	fmt.Printf("✅ Found %d occurrences\n", foundCount)

	missingKeys := []string{}
	for key := range keys {
		if _, exists := foundKeys[key]; !exists {
			missingKeys = append(missingKeys, key)
		}
	}

	// Write missing keys to a file
	fmt.Printf("\n❌ MISSING KEYS: %d\n", len(missingKeys))

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
