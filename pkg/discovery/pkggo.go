package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type Dependent struct {
	ImportPath string `json:"import_path"`
	Synopsis   string `json:"synopsis"`
}

func FetchDependents(module string, limit int) ([]Dependent, error) {
	if limit <= 0 {
		limit = 10
	}

	escapedModule := url.QueryEscape(module)
	pkgURL := fmt.Sprintf("https://pkg.go.dev/%s?tab=importedby", escapedModule)

	resp, err := http.Get(pkgURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("pkg.go.dev returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	dependents := parseDependentsFromHTML(string(body), limit)
	
	if len(dependents) == 0 {
		return nil, fmt.Errorf("no dependents found for %s", module)
	}

	return dependents, nil
}

func parseDependentsFromHTML(html string, limit int) []Dependent {
	var dependents []Dependent
	
	pathRegex := regexp.MustCompile(`<a href="/(github\.com/[^"]+)"[^>]*>`)
	matches := pathRegex.FindAllStringSubmatch(html, -1)
	
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(dependents) >= limit {
			break
		}
		
		if len(match) > 1 {
			path := match[1]
			if !seen[path] && !strings.Contains(path, "?") {
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