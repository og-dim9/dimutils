package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dimutils",
	Short: "Dim9 utilities multicall binary",
	Long:  `A multicall binary containing various data processing and transformation utilities.`,
}

func main() {
	// Get the name this binary was called with
	progName := filepath.Base(os.Args[0])
	
	// Remove common executable extensions
	progName = strings.TrimSuffix(progName, ".exe")
	
	// If called as dimutils, show help or execute subcommand
	if progName == "dimutils" {
		if err := rootCmd.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	
	// Otherwise, find and execute the matching subcommand
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == progName || contains(cmd.Aliases, progName) {
			// Prepend the command name to args and execute
			newArgs := append([]string{os.Args[0], progName}, os.Args[1:]...)
			os.Args = newArgs
			if err := rootCmd.Execute(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}
	
	// If we get here, the program name didn't match any command
	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", progName)
	os.Exit(1)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func init() {
	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("dimutils version 0.3.0")
		},
	})
}