package main

import (
	"bbrew/internal/services"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// Define flags
	brewfilePath := flag.String("f", "", "Path to Brewfile (show only packages from this Brewfile)")
	showVersion := flag.Bool("v", false, "Show version information")
	flag.Bool("version", false, "Show version information")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Bold Brew - A TUI for Homebrew package management\n\n")
		fmt.Fprintf(os.Stderr, "Usage: bbrew [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -f <path|url> Path or URL to Brewfile\n")
		fmt.Fprintf(os.Stderr, "  -v, --version Show version information\n")
		fmt.Fprintf(os.Stderr, "  -h, --help    Show this help message\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  bbrew                    Launch the TUI with all packages\n")
		fmt.Fprintf(os.Stderr, "  bbrew -f ~/Brewfile      Launch with packages from local Brewfile\n")
		fmt.Fprintf(os.Stderr, "  bbrew -f https://...     Launch with packages from remote Brewfile\n")
	}

	flag.Parse()

	// Handle --version flag (check both -v and --version)
	if *showVersion || isFlagPassed("version") {
		fmt.Printf("Bold Brew %s\n", services.AppVersion)
		os.Exit(0)
	}

	// Resolve Brewfile path (handles both local and remote URLs)
	var cleanup func()
	if *brewfilePath != "" {
		localPath, cleanupFn, err := services.ResolveBrewfilePath(*brewfilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		*brewfilePath = localPath
		cleanup = cleanupFn
		defer cleanup()
	}

	// Initialize app service
	var appService services.AppServiceInterface = services.NewAppService()

	// Ensure cleanup runs on exit (if appService implements Cleanup)
	if s, ok := appService.(*services.AppService); ok {
		defer s.Cleanup()
	}

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

// isFlagPassed checks if a flag was explicitly passed on the command line.
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
