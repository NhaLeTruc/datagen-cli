// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/NhaLeTruc/datagen-cli/internal/cli"
	"github.com/spf13/cobra/doc"
)

func main() {
	// Get the output directory from command line args or use default
	outputDir := "docs/man"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create man directory: %v", err)
	}

	// Get the absolute path
	absPath, err := filepath.Abs(outputDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create root command
	rootCmd := cli.NewRootCommand()

	// Generate man pages for all commands
	header := &doc.GenManHeader{
		Title:   "DATAGEN",
		Section: "1",
		Source:  "datagen-cli",
		Manual:  "User Commands",
	}

	if err := doc.GenManTree(rootCmd, header, absPath); err != nil {
		log.Fatalf("Failed to generate man pages: %v", err)
	}

	fmt.Printf("✓ Man pages generated successfully in %s\n", absPath)

	// List generated files
	files, err := os.ReadDir(absPath)
	if err != nil {
		log.Fatalf("Failed to read man directory: %v", err)
	}

	fmt.Println("\nGenerated man pages:")
	for _, file := range files {
		if !file.IsDir() {
			fmt.Printf("  - %s\n", file.Name())
		}
	}
}
