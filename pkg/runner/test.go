package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"deptest/pkg/discovery"
)

type TestResult struct {
	Project string `json:"project"`
	Status  string `json:"status"`
	Details string `json:"details"`
	Duration float64 `json:"duration"`
}

func RunTests(deps []discovery.Dependent, workDir string, timeout time.Duration) ([]TestResult, error) {
	if workDir == "" {
		workDir = "workspace"
	}
	
	err := os.MkdirAll(workDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}
	
	var results []TestResult
	
	for i, dep := range deps {
		fmt.Printf("[%d/%d] Testing %s...\n", i+1, len(deps), dep.ImportPath)
		
		result := runSingleTest(dep, workDir, timeout)
		results = append(results, result)
		
		fmt.Printf("  Status: %s\n", result.Status)
	}
	
	return results, nil
}

func runSingleTest(dep discovery.Dependent, workDir string, timeout time.Duration) TestResult {
	start := time.Now()
	
	repoURL := fmt.Sprintf("https://%s.git", dep.ImportPath)
	repoName := filepath.Base(dep.ImportPath)
	repoPath := filepath.Join(workDir, repoName)
	
	if _, err := os.Stat(repoPath); err == nil {
		os.RemoveAll(repoPath)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	cloneCmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", repoURL, repoPath)
	cloneCmd.Stdout = io.Discard
	cloneCmd.Stderr = io.Discard
	
	err := cloneCmd.Run()
	if err != nil {
		duration := time.Since(start).Seconds()
		return TestResult{
			Project: dep.ImportPath,
			Status: "error",
			Details: fmt.Sprintf("clone failed: %v", err),
			Duration: duration,
		}
	}
	
	testCtx, testCancel := context.WithTimeout(context.Background(), timeout)
	defer testCancel()
	
	testCmd := exec.CommandContext(testCtx, "go", "test", "./...")
	testCmd.Dir = repoPath
	output, err := testCmd.CombinedOutput()
	
	duration := time.Since(start).Seconds()
	
	if err != nil {
		if testCtx.Err() == context.DeadlineExceeded {
			return TestResult{
				Project: dep.ImportPath,
				Status: "timeout",
				Details: "test execution exceeded timeout",
				Duration: duration,
			}
		}
		
		return TestResult{
			Project: dep.ImportPath,
			Status: "fail",
			Details: extractErrorSummary(string(output)),
			Duration: duration,
		}
	}
	
	return TestResult{
		Project: dep.ImportPath,
		Status: "pass",
		Details: "all tests passed",
		Duration: duration,
	}
}

func extractErrorSummary(output string) string {
	lines := strings.Split(output, "\n")
	var errorLines []string
	
	for _, line := range lines {
		if strings.Contains(line, "FAIL") || strings.Contains(line, "Error") {
			errorLines = append(errorLines, line)
			if len(errorLines) >= 3 {
				break
			}
		}
	}
	
	if len(errorLines) > 0 {
		return strings.Join(errorLines, "; ")
	}
	
	if len(output) > 200 {
		return output[:200] + "..."
	}
	
	return output
}

func SaveResults(results []TestResult, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
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

func LoadResults(filename string) ([]TestResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	var results []TestResult
	err = json.Unmarshal(data, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	return results, nil
}