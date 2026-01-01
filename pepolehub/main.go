package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	name := flag.String("name", "", "Person's name")
	linkedin := flag.String("linkedin", "", "LinkedIn URL")
	flag.Parse()

	if *name == "" || *linkedin == "" {
		fmt.Println("Usage: pepolehub -name \"Name\" -linkedin \"URL\"")
		os.Exit(1)
	}

	fmt.Printf("🚀 Starting PeopleHub Research for: %s (%s)\n", *name, *linkedin)

	// Initialize Graph
	runnable, err := NewResearchGraph()
	if err != nil {
		log.Fatalf("Failed to create graph: %v", err)
	}

	// Initial State
	initialState := ResearchState{
		PersonName:  *name,
		LinkedinUrl: *linkedin,
	}

	// Run
	ctx := context.Background()
	finalState, err := runnable.Invoke(ctx, initialState)
	if err != nil {
		log.Fatalf("Graph execution failed: %v", err)
	}

	// Output
	fmt.Println("\n✅ Research Complete!")
	fmt.Println("--------------------------------------------------")
	if finalState.FinalReport != "" {
		fmt.Println(finalState.FinalReport)
	} else {
		fmt.Println("No report generated.")
		if len(finalState.Errors) > 0 {
			fmt.Println("Errors encountered:")
			for _, e := range finalState.Errors {
				fmt.Println("- ", e)
			}
		}
	}
	fmt.Println("--------------------------------------------------")
}
