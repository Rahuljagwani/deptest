package compare

import (
	"fmt"
	"deptest/pkg/runner"
)

type ComparisonResult struct {
	TotalProjects int
	NewlyBroken   []string
	NewlyFixed    []string
	StillPassing  []string
	StillFailing  []string
	StatusChanged int
}

func Compare(before, after []runner.TestResult) ComparisonResult {
	result := ComparisonResult{
		TotalProjects: len(before),
	}
	
	afterMap := make(map[string]runner.TestResult)
	for _, r := range after {
		afterMap[r.Project] = r
	}
	
	for _, beforeRes := range before {
		afterRes, exists := afterMap[beforeRes.Project]
		if !exists {
			continue
		}
		
		beforePassed := beforeRes.Status == "pass"
		afterPassed := afterRes.Status == "pass"
		
		if beforePassed && !afterPassed {
			result.NewlyBroken = append(result.NewlyBroken, beforeRes.Project)
			result.StatusChanged++
		} else if !beforePassed && afterPassed {
			result.NewlyFixed = append(result.NewlyFixed, beforeRes.Project)
			result.StatusChanged++
		} else if beforePassed && afterPassed {
			result.StillPassing = append(result.StillPassing, beforeRes.Project)
		} else {
			result.StillFailing = append(result.StillFailing, beforeRes.Project)
		}
	}
	
	return result
}

func PrintComparison(result ComparisonResult) {
	fmt.Println("\nImpact Analysis")
	fmt.Printf("Total projects analyzed: %d\n", result.TotalProjects)
	fmt.Printf("Status changes: %d\n\n", result.StatusChanged)
	
	if len(result.NewlyBroken) > 0 {
		fmt.Printf("NEWLY BROKEN (%d):\n", len(result.NewlyBroken))
		for _, proj := range result.NewlyBroken {
			fmt.Printf("  - %s\n", proj)
		}
		fmt.Println()
	}
	
	if len(result.NewlyFixed) > 0 {
		fmt.Printf("NEWLY FIXED (%d):\n", len(result.NewlyFixed))
		for _, proj := range result.NewlyFixed {
			fmt.Printf("  - %s\n", proj)
		}
		fmt.Println()
	}
	
	fmt.Printf("Still passing: %d\n", len(result.StillPassing))
	fmt.Printf("Still failing: %d\n", len(result.StillFailing))
	
	if len(result.NewlyBroken) > 0 {
		fmt.Println("\nThis change breaks existing dependents!")
	} else if result.StatusChanged == 0 {
		fmt.Println("\nNo impact detected - change appears safe!")
	}
}