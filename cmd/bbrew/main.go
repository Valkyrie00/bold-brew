package main

import (
	"bbrew/internal/services"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// Parse command line flags
	brewfilePath := flag.String("f", "", "Path to Brewfile (show only packages from this Brewfile)")
	flag.Parse()

	// Validate Brewfile path if provided
	if *brewfilePath != "" {
		if _, err := os.Stat(*brewfilePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Brewfile not found: %s\n", *brewfilePath)
			os.Exit(1)
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Cannot access Brewfile: %v\n", err)
			os.Exit(1)
		}
	}

	// Initialize app service
	appService := services.NewAppService()
	// Configure Brewfile mode if path was provided
	if *brewfilePath != "" {
		appService.SetBrewfilePath(*brewfilePath)
	}

	// Boot the application (load Homebrew data)
	if err := appService.Boot(); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Build and run the TUI
	appService.BuildApp()
	if err := appService.GetApp().Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
