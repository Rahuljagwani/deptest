package main

import (
	"fmt"
	"os"
	"time"
	"deptest/pkg/discovery"
	"deptest/pkg/runner"
)

func main() {
	handleDiscover()
	handleTest()
}

func handleDiscover() {

	// hardcode for now
	module := "github.com/sirupsen/logrus"
	limit := 10
	output := "dependents.json"

	for i := 3; i < len(os.Args); i++ {
		if os.Args[i] == "-limit" && i+1 < len(os.Args) {
			fmt.Sscanf(os.Args[i+1], "%d", &limit)
			i++
		} else if os.Args[i] == "-o" && i+1 < len(os.Args) {
			output = os.Args[i+1]
			i++
		}
	}

	fmt.Printf("Discovering dependents for %s (limit: %d)...\n", module, limit)

	deps, err := discovery.FetchDependents(module, limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d dependents\n", len(deps))
	
	for i, dep := range deps {
		fmt.Printf("  %d. %s\n", i+1, dep.ImportPath)
	}

	err = discovery.SaveDependents(deps, output)
	if err != nil {
		fmt.Printf("Error saving results: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSaved to %s\n", output)
}

func handleTest() {
	input := ""
	output := "results.json"
	timeout := 2 * time.Minute

	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-i" && i+1 < len(os.Args) {
			input = os.Args[i+1]
			i++
		} else if os.Args[i] == "-o" && i+1 < len(os.Args) {
			output = os.Args[i+1]
			i++
		} else if os.Args[i] == "-timeout" && i+1 < len(os.Args) {
			d, err := time.ParseDuration(os.Args[i+1])
			if err == nil {
				timeout = d
			}
			i++
		}
	}

	if input == "" {
		fmt.Println("Error: input file required")
		fmt.Println("Usage: deptest test -i dependents.json -o results.json")
		os.Exit(1)
	}

	deps, err := discovery.LoadDependents(input)
	if err != nil {
		fmt.Printf("Error loading dependents: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Running tests for %d projects (timeout: %v per project)...\n\n", len(deps), timeout)

	results, err := runner.RunTests(deps, "workspace", timeout)
	if err != nil {
		fmt.Printf("Error running tests: %v\n", err)
		os.Exit(1)
	}

	err = runner.SaveResults(results, output)
	if err != nil {
		fmt.Printf("Error saving results: %v\n", err)
		os.Exit(1)
	}

	passed := 0
	failed := 0
	errors := 0
	
	for _, r := range results {
		switch r.Status {
		case "pass":
			passed++
		case "fail":
			failed++
		default:
			errors++
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Passed: %d\n", passed)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Errors: %d\n", errors)
	fmt.Printf("\nResults saved to %s\n", output)
}