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

	// Print configuration
	if config.Verbose {
		fmt.Println(config.String())
		fmt.Println()
	}

	// Get research query from command line or use default
	query := "What are the latest advances in large language models in 2024?"
	if len(os.Args) > 1 {
		query = os.Args[1]
	}

	// Create researcher
	researcher, err := NewGPTResearcher(config)
	if err != nil {
		log.Fatalf("Failed to create researcher: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Conduct research
	state, err := researcher.ConductResearch(ctx, query)
	if err != nil {
		log.Fatalf("Research failed: %v", err)
	}

	// Print final report
	if config.Verbose {
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println("FINAL RESEARCH REPORT")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println()
	}

	fmt.Println(state.FinalReport)

	// Optionally save to file
	if config.SaveIntermediate {
		filename := fmt.Sprintf("%s/research_report_%s.md",
			config.OutputDir,
			time.Now().Format("20060102_150405"))

		if err := researcher.PublisherAgent.SaveReport(state, filename); err != nil {
			log.Printf("Failed to save report: %v", err)
		}
	}
}
