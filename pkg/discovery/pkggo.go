package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type Dependent struct {
	ImportPath string `json:"import_path"`
	Synopsis   string `json:"synopsis"`
}

func FetchDependents(module string, limit int) ([]Dependent, error) {
	if limit <= 0 {
		limit = 10
	}

	pkgURL := fmt.Sprintf("https://pkg.go.dev/%s?tab=importedby", module)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(pkgURL)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("pkg.go.dev returned status %d for %s", resp.StatusCode, pkgURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	dependents := parseDependentsFromHTML(string(body), limit)

	println("dependents found", len(dependents))

	if len(dependents) == 0 {
		return nil, fmt.Errorf("no dependents found for module %s", module)
	}

	return dependents, nil
}

func parseDependentsFromHTML(html string, limit int) []Dependent {
	var dependents []Dependent
	
	linkRegex := regexp.MustCompile(`<a[^>]+href="/((?:github\.com|gitlab\.com|bitbucket\.org)/[^"?]+)"`)
	matches := linkRegex.FindAllStringSubmatch(html, -1)
	
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(dependents) >= limit {
			break
		}
		
		if len(match) > 1 {
			path := match[1]
			if !seen[path] && isValidDependency(path) {
				seen[path] = true
				dependents = append(dependents, Dependent{
					ImportPath: path,
					Synopsis:   "",
				})
			}
		}
	}
	
	return dependents
}

func isValidDependency(path string) bool {
	if strings.Contains(path, "?") || strings.Contains(path, "#") {
		return false
	}
	
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return false
	}
	
	return true
}

func SaveDependents(deps []Dependent, filename string) error {
	data, err := json.MarshalIndent(deps, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

func LoadDependents(filename string) ([]Dependent, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	var deps []Dependent
	err = json.Unmarshal(data, &deps)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	return deps, nil
}