package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	// Load configuration
	config := NewConfig()

	// Print configuration if verbose
	if config.Verbose {
		fmt.Println(config.String())
		fmt.Println()
	}

	// Get query from command line or use default
	query := "Research the latest advances in AI agents and create a summary report with key findings"
	if len(os.Args) > 1 {
		query = strings.Join(os.Args[1:], " ")
	}

	// Create LangManus instance
	lm, err := NewLangManus(config)
	if err != nil {
		log.Fatalf("Failed to create LangManus: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Run the workflow
	state, err := lm.Run(ctx, query)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	// Print final report
	printFinalReport(state)
}

func printFinalReport(state *State) {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("                         FINAL REPORT")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	fmt.Printf("Query: %s\n\n", state.Query)

	if state.FinalReport != "" {
		fmt.Println(state.FinalReport)
	} else {
		fmt.Println("No final report generated.")
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("                         EXECUTION SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Print agent history
	fmt.Printf("Agents involved: ")
	agentSet := make(map[AgentType]bool)
	for _, agent := range state.AgentHistory {
		agentSet[agent] = true
	}
	agents := []string{}
	for agent := range agentSet {
		agents = append(agents, string(agent))
	}
	fmt.Println(strings.Join(agents, ", "))
	fmt.Println()

	// Print research results
	if len(state.ResearchResults) > 0 {
		fmt.Printf("Research conducted: %d searches\n", len(state.ResearchResults))
		for i, result := range state.ResearchResults {
			fmt.Printf("  %d. %s (%d sources)\n", i+1, result.Query, len(result.Sources))
		}
		fmt.Println()
	}

	// Print code executions
	if len(state.CodeResults) > 0 {
		fmt.Printf("Code executed: %d times\n", len(state.CodeResults))
		for i, result := range state.CodeResults {
			status := "✓"
			if result.Error != "" {
				status = "✗"
			}
			fmt.Printf("  %s %d. Exit code: %d\n", status, i+1, result.ExitCode)
		}
		fmt.Println()
	}

	// Print tasks
	fmt.Printf("Tasks completed: %d/%d\n", len(state.GetCompletedTasks()), len(state.Tasks))
	fmt.Println()

	fmt.Printf("Total messages: %d\n", len(state.Messages))
	fmt.Printf("Status: %s\n", state.Status)
	fmt.Println()

	fmt.Println(strings.Repeat("=", 80))
}
