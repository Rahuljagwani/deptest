package main

import (
	"fmt"
	"os"
	"deptest/pkg/discovery"
)

func main() {
	handleDiscover()
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